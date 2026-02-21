// pkg/deploy/ssh.go — SSH auto-deploy for remote nodes.
package deploy

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	gossh "golang.org/x/crypto/ssh"

	"github.com/Chinsusu/proxy-server-local/pkg/types"
)

// MasterURL is the URL of this PGW master server, set at startup.
var MasterURL string

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

// DeployNode SSHes into node and runs the install script.
// logChan receives log lines (closed when done). ctx controls cancellation.
func DeployNode(ctx context.Context, node types.Node, secret string, logChan chan<- string) error {
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
			return fmt.Errorf("parse ssh key: %w", err)
		}
		auths = append(auths, gossh.PublicKeys(signer))
	}
	if pass != "" {
		auths = append(auths, gossh.Password(pass))
	}
	if len(auths) == 0 {
		return fmt.Errorf("no SSH credentials configured")
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
		return fmt.Errorf("dial %s: %w", addr, err)
	}

	sshConn, chans, reqs, err := gossh.NewClientConn(conn, addr, cfg)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("ssh handshake: %w", err)
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

		out := &strings.Builder{}
		sess.Stdout = out
		sess.Stderr = out

		log(fmt.Sprintf("[deploy] $ %s", cmd))
		if err := sess.Run(cmd); err != nil {
			log(fmt.Sprintf("[deploy] ERROR: %v\n%s", err, out.String()))
			return fmt.Errorf("run %q: %w", cmd, err)
		}
		if s := strings.TrimSpace(out.String()); s != "" {
			log(s)
		}
		return nil
	}

	// Step 1: Install dependencies
	log("[deploy] Step 1/5: apt update + install curl systemd ...")
	if err := run("DEBIAN_FRONTEND=noninteractive apt-get update -qq && apt-get install -y -qq curl systemd"); err != nil {
		return err
	}

	// Step 2: Download and install pgw binaries from GitHub releases
	log("[deploy] Step 2/5: Install pgw-node binaries ...")
	installScript := fmt.Sprintf(`
set -e
REPO="Chinsusu/pgw-node"
BIN_URL="https://github.com/$REPO/releases/latest/download"
mkdir -p /usr/local/bin /etc/pgw /etc/pgw-node /var/lib/pgw

# Install pgw-node
curl -fsSL "$BIN_URL/pgw-node-linux-amd64" -o /usr/local/bin/pgw-node && chmod +x /usr/local/bin/pgw-node

# Install proxy-server-local binaries (api, agent, fwd, health)
MASTER_REPO="Chinsusu/proxy-server-local"
MASTER_URL="https://github.com/$MASTER_REPO/releases/latest/download"
for bin in pgw-api pgw-agent pgw-fwd pgw-health; do
  curl -fsSL "$MASTER_URL/$bin-linux-amd64" -o /usr/local/bin/$bin && chmod +x /usr/local/bin/$bin
done
echo "Binaries installed"
`)
	if err := run(strings.TrimSpace(installScript)); err != nil {
		return err
	}

	// Step 3: Write environment files
	log("[deploy] Step 3/5: Writing config files ...")
	masterURL := MasterURL
	if masterURL == "" {
		masterURL = "http://master:8080"
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
		return err
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

# pgw-node has its own env file
sed -i 's|EnvironmentFile=/etc/pgw/pgw.env|EnvironmentFile=/etc/pgw-node/pgw-node.env|' /etc/systemd/system/pgw-node.service
systemctl daemon-reload
`
	if err := run(strings.TrimSpace(unitInstall)); err != nil {
		return err
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
		return fmt.Errorf("session for step 5: %w", err)
	}
	out := &strings.Builder{}
	sess.Stdout = out
	sess.Stderr = out
	log(fmt.Sprintf("[deploy] $ %s", "pgw-node init + start services"))
	_ = sess.Run(strings.TrimSpace(startCmd))
	sess.Close()

	result := out.String()
	log(result)

	// Try to extract public key from output and update node record
	for _, line := range strings.Split(result, "\n") {
		if strings.HasPrefix(line, "PUBKEY:") {
			pubkey := strings.TrimPrefix(line, "PUBKEY:")
			pubkey = strings.TrimSpace(pubkey)
			if pubkey != "" {
				log(fmt.Sprintf("[deploy] Node public key: %s", pubkey))
			}
		}
	}

	log("[deploy] Deploy completed successfully ✓")
	return nil
}
