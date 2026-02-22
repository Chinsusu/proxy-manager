// pkg/deploy/ssh.go — SSH auto-deploy for remote nodes.
package deploy

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	gossh "golang.org/x/crypto/ssh"

	"github.com/Chinsusu/proxy-server-local/pkg/types"
)

// MasterURL is the URL of this PGW master server, set at startup.
var MasterURL string

// NodeBinSourceDir is the directory containing the pgw-node Go module source.
// When set, the master will build pgw-node locally and transfer the binary via SSH
// instead of downloading from GitHub releases.
var NodeBinSourceDir string

// detectOutboundIP returns the local IP used to reach external hosts.
// This is used to replace 127.0.0.1/localhost in MasterURL for node config,
// so remote VPS nodes can actually reach the master server.
func detectOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return ""
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// Encrypt encrypts plaintext using AES-256-GCM, key derived from secret.
// Returns hex-encoded ciphertext.
func Encrypt(plaintext, secret string) (string, error) {
	key := deriveKey(secret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ct), nil
}

// Decrypt decrypts hex-encoded AES-256-GCM ciphertext.
func Decrypt(hexCT, secret string) (string, error) {
	if hexCT == "" {
		return "", nil
	}
	ct, err := hex.DecodeString(hexCT)
	if err != nil {
		return "", err
	}
	key := deriveKey(secret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ns := gcm.NonceSize()
	if len(ct) < ns {
		return "", fmt.Errorf("ciphertext too short")
	}
	pt, err := gcm.Open(nil, ct[:ns], ct[ns:], nil)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

func deriveKey(secret string) []byte {
	h := sha256.Sum256([]byte(secret))
	return h[:]
}

// buildLocalNodeBin cross-compiles pgw-node for linux/amd64 from NodeBinSourceDir.
// Returns path to the temp binary or error.
func buildLocalNodeBin() (string, error) {
	// Locate go binary — systemd services may not have it in PATH.
	goBin, err := exec.LookPath("go")
	if err != nil {
		for _, p := range []string{"/usr/local/go/bin/go", "/usr/bin/go", "/usr/local/bin/go"} {
			if _, e := os.Stat(p); e == nil {
				goBin = p
				break
			}
		}
	}
	if goBin == "" {
		return "", fmt.Errorf("go binary not found; install Go on the master server")
	}

	tmp, err := os.CreateTemp("", "pgw-node-*")
	if err != nil {
		return "", err
	}
	tmp.Close()
	cmd := exec.Command(goBin, "build", "-buildvcs=false", "-o", tmp.Name(), "./cmd/pgw-node")
	cmd.Dir = NodeBinSourceDir
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0",
		"PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin",
		"HOME=/root",
		"GOCACHE=/tmp/pgw-go-cache",
		"GOPATH=/tmp/pgw-go-path")
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = os.Remove(tmp.Name())
		return "", fmt.Errorf("build pgw-node: %w\n%s", err, out)
	}
	return tmp.Name(), nil
}

// transferBinary sends a local file to remotePath by base64-encoding it and
// streaming through SSH stdin → base64 -d. No scp/sftp binary required on remote.
func transferBinary(client *gossh.Client, localPath, remotePath string) error {
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("read local binary: %w", err)
	}

	// Ensure parent directory exists first
	mkdirSess, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("mkdir session: %w", err)
	}
	dir := "/"
	for i := len(remotePath) - 1; i >= 0; i-- {
		if remotePath[i] == '/' {
			dir = remotePath[:i]
			break
		}
	}
	_ = mkdirSess.Run("mkdir -p " + dir)
	mkdirSess.Close()

	// Encode binary to base64 and pipe through stdin
	encoded := base64.StdEncoding.EncodeToString(data)
	sess, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("transfer session: %w", err)
	}
	defer sess.Close()

	var stderr bytes.Buffer
	sess.Stderr = &stderr
	sess.Stdin = strings.NewReader(encoded + "\n")

	cmd := fmt.Sprintf("base64 -d > %s && chmod +x %s", remotePath, remotePath)
	if err := sess.Run(cmd); err != nil {
		return fmt.Errorf("base64 transfer: %w (stderr: %s)", err, stderr.String())
	}
	return nil
}

// DeployNode SSHes into node and runs the install script.
// logChan receives log lines (closed when done). ctx controls cancellation.
// Returns the node's Ed25519 public key (hex) extracted after init, or empty if unavailable.
func DeployNode(ctx context.Context, node types.Node, secret string, logChan chan<- string) (string, error) {
	defer close(logChan)

	log := func(msg string) {
		select {
		case logChan <- msg:
		case <-ctx.Done():
		}
	}

	// Decrypt SSH credentials
	pass, _ := Decrypt(node.SSHPassword, secret)
	key, _ := Decrypt(node.SSHKey, secret)

	// Build SSH auth methods
	var auths []gossh.AuthMethod
	if key != "" {
		signer, err := gossh.ParsePrivateKey([]byte(key))
		if err != nil {
			return "", fmt.Errorf("parse ssh key: %w", err)
		}
		auths = append(auths, gossh.PublicKeys(signer))
	}
	if pass != "" {
		auths = append(auths, gossh.Password(pass))
	}
	if len(auths) == 0 {
		return "", fmt.Errorf("no SSH credentials configured")
	}

	host := node.SSHHost
	port := node.SSHPort
	if port == 0 {
		port = 22
	}
	addr := fmt.Sprintf("%s:%d", host, port)

	log(fmt.Sprintf("[deploy] Connecting to %s@%s ...", node.SSHUser, addr))

	cfg := &gossh.ClientConfig{
		User:            node.SSHUser,
		Auth:            auths,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), // nolint
		Timeout:         30 * time.Second,
	}

	dialer := &net.Dialer{Timeout: 30 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return "", fmt.Errorf("dial %s: %w", addr, err)
	}

	sshConn, chans, reqs, err := gossh.NewClientConn(conn, addr, cfg)
	if err != nil {
		_ = conn.Close()
		return "", fmt.Errorf("ssh handshake: %w", err)
	}
	client := gossh.NewClient(sshConn, chans, reqs)
	defer client.Close()
	log("[deploy] Connected ✓")

	run := func(cmd string) error {
		sess, err := client.NewSession()
		if err != nil {
			return err
		}
		defer sess.Close()

		// Stream stdout+stderr real-time, line by line.
		pr, pw := io.Pipe()
		sess.Stdout = pw
		sess.Stderr = pw

		log(fmt.Sprintf("[deploy] $ %s", cmd))

		var runErr error
		go func() {
			runErr = sess.Run(cmd)
			pw.Close()
		}()

		buf := make([]byte, 0, 256)
		tmp := make([]byte, 512)
		for {
			n, err := pr.Read(tmp)
			if n > 0 {
				buf = append(buf, tmp[:n]...)
				for {
					i := bytes.IndexByte(buf, '\n')
					if i < 0 { break }
					log(string(buf[:i]))
					buf = buf[i+1:]
				}
			}
			if err != nil { break } // io.EOF or pipe closed
		}
		if len(buf) > 0 {
			log(string(buf))
		}
		if runErr != nil {
			log(fmt.Sprintf("[deploy] ERROR: %v", runErr))
			return fmt.Errorf("run %q: %w", cmd, runErr)
		}
		return nil
	}

	// Step 1: Install dependencies
	log("[deploy] Step 1/5: apt update + install curl systemd ...")
	apt1 := "DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a NEEDRESTART_SUSPEND=1 " +
		"apt-get update -qq && apt-get install -y -qq curl systemd"
	if err := run(apt1); err != nil {
		return "", err
	}

	// Step 2: Install pgw-node binary (local build transfer or GitHub release)
	log("[deploy] Step 2/5: Install pgw-node binary ...")
	if NodeBinSourceDir != "" {
		// Build locally and transfer via SSH — no GitHub release required.
		log("[deploy] Building pgw-node from source ...")
		binPath, err := buildLocalNodeBin()
		if err != nil {
			log(fmt.Sprintf("[deploy] WARN: local build failed: %v", err))
			log("[deploy] Falling back to GitHub release download ...")
			if err2 := run(`
set -e
mkdir -p /usr/local/bin
curl -fsSL "https://github.com/Chinsusu/pgw-node/releases/latest/download/pgw-node-linux-amd64" -o /usr/local/bin/pgw-node && chmod +x /usr/local/bin/pgw-node
echo "pgw-node installed from GitHub"
`); err2 != nil {
				return "", err2
			}
		} else {
			defer os.Remove(binPath)
			log("[deploy] Transferring pgw-node binary via SSH ...")
			if err := transferBinary(client, binPath, "/usr/local/bin/pgw-node"); err != nil {
				return "", fmt.Errorf("transfer pgw-node binary: %w", err)
			}
			log("[deploy] pgw-node binary transferred ✓")
		}
	} else {
		// Download from GitHub releases.
		installScript := `
set -e
mkdir -p /usr/local/bin
curl -fsSL "https://github.com/Chinsusu/pgw-node/releases/latest/download/pgw-node-linux-amd64" -o /usr/local/bin/pgw-node && chmod +x /usr/local/bin/pgw-node
echo "pgw-node installed from GitHub"
`
		if err := run(strings.TrimSpace(installScript)); err != nil {
			return "", err
		}
	}

	// Step 3: Write environment files
	log("[deploy] Step 3/5: Writing config files ...")
	masterURL := MasterURL
	if masterURL == "" {
		masterURL = "http://master:8080"
	}
	// Node VPS cannot reach localhost of the master — replace with outbound IP.
	if strings.Contains(masterURL, "127.0.0.1") || strings.Contains(masterURL, "localhost") {
		if ip := detectOutboundIP(); ip != "" {
			masterURL = strings.NewReplacer("127.0.0.1", ip, "localhost", ip).Replace(masterURL)
			log("[deploy] Master IP for node config: " + masterURL)
		}
	}
	pgwEnv := fmt.Sprintf(`# PGW Node config (auto-generated by master deploy)
PGW_STORE=sqlite
PGW_STORE_PATH=/var/lib/pgw/state.db
PGW_API_ADDR=:8080
PGW_HEALTH_INTERVAL=30s
PGW_STRICT_OUTPUT=true
PGW_WAN_IFACE=eth0
PGW_LAN_IFACE=eth0
`)
	pgwNodeEnv := fmt.Sprintf(`# pgw-node sync daemon config
PGW_SERVER=%s
PGW_NODE_ID=%s
PGW_KEY_PATH=/etc/pgw-node/node.key
PGW_POLL_INTERVAL=15s
`, masterURL, node.ID)

	writeCmd := fmt.Sprintf(
		`mkdir -p /etc/pgw /etc/pgw-node
cat > /etc/pgw/pgw.env << 'ENVEOF'
%s
ENVEOF
cat > /etc/pgw-node/pgw-node.env << 'ENVEOF'
%s
ENVEOF
chmod 600 /etc/pgw/pgw.env /etc/pgw-node/pgw-node.env`,
		pgwEnv, pgwNodeEnv,
	)
	if err := run(writeCmd); err != nil {
		return "", err
	}

	// Step 4: Install systemd units
	log("[deploy] Step 4/5: Installing systemd units ...")
	unitInstall := `
for svc in pgw-api pgw-agent pgw-health pgw-node; do
  cat > /etc/systemd/system/$svc.service << EOF
[Unit]
Description=PGW $svc
After=network.target

[Service]
EnvironmentFile=/etc/pgw/pgw.env
ExecStart=/usr/local/bin/$svc
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF
done

# pgw-node has its own env file AND needs 'dev' subcommand
sed -i 's|EnvironmentFile=/etc/pgw/pgw.env|EnvironmentFile=/etc/pgw-node/pgw-node.env|' /etc/systemd/system/pgw-node.service
sed -i 's|ExecStart=/usr/local/bin/pgw-node$|ExecStart=/usr/local/bin/pgw-node dev|' /etc/systemd/system/pgw-node.service
systemctl daemon-reload
`
	if err := run(strings.TrimSpace(unitInstall)); err != nil {
		return "", err
	}

	// Step 5: Initialize pgw-node keypair and start services
	log("[deploy] Step 5/5: Init keypair + start services ...")
	startCmd := `
set -e
# Generate keypair if not exists
if [ ! -f /etc/pgw-node/node.key ]; then
  /usr/local/bin/pgw-node init
fi
# Enable + start all services
systemctl enable --now pgw-api pgw-agent pgw-health pgw-node 2>&1 || true
# Print public key for registration
PUBKEY=$(PGW_KEY_PATH=/etc/pgw-node/node.key /usr/local/bin/pgw-node pubkey 2>/dev/null || cat /etc/pgw-node/node.pub 2>/dev/null || echo "")
echo "PUBKEY:$PUBKEY"
`
	sess, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("session for step 5: %w", err)
	}
	out := &strings.Builder{}
	sess.Stdout = out
	sess.Stderr = out
	log(fmt.Sprintf("[deploy] $ %s", "pgw-node init + start services"))
	_ = sess.Run(strings.TrimSpace(startCmd))
	sess.Close()

	result := out.String()
	log(result)

	// Extract Ed25519 public key from output (line: "PUBKEY:<hex>")
	var pubkey string
	for _, line := range strings.Split(result, "\n") {
		if strings.HasPrefix(line, "PUBKEY:") {
			pubkey = strings.TrimSpace(strings.TrimPrefix(line, "PUBKEY:"))
			if pubkey != "" {
				log(fmt.Sprintf("[deploy] ✓ Node public key captured: %s", pubkey))
			}
			break
		}
	}

	log("[deploy] Deploy completed successfully ✓")
	return pubkey, nil
}
