# セキュアWOL/Shutdownマイクロサービス (pve-wol)

## 概要
* **ホスト環境**: ProxmoxVE上の独立したLXC (Alpine Linux)
* **目的**: プライベートネットワーク内のPCをセキュアにWake-on-LANおよびシャットダウンするAPIサーバー。
## API仕様
* **ベースURL**: `http://10.0.0.X:8080` (LXCのvnet0側IP)
* **エンドポイント**:
    * `POST /wake`
        * **処理**: 物理PCのMACアドレス宛にWOLマジックパケットをブロードキャストする。環境変数で指定したブロードキャストアドレスを使用する。
    * `POST /shutdown`
        * **処理**: Go内部のSSHクライアント (`golang.org/x/crypto/ssh`) を利用し、物理PC (Windows) へ接続して `shutdown /s /t 3` を実行する。
    * `GET /status`
        * **処理**: Go内部実装のICMP Pingを対象IPへ送信し、応答があれば `ONLINE` (200 OK)、なければ `OFFLINE` (503 Service Unavailable) を返す。

## リリース・転送方法
ビルドしたバイナリをLXCへ転送する方法は以下の2通りがあります。

### GitHub Actionsを使用する（推奨）
タグをプッシュすると自動的にビルドされ、GitHubのReleasesにバイナリがアップロードされます。
```bash
git tag v1.0.0
git push origin v1.0.0
```
LXCコンソールに入り、Releasesから直接ダウンロードします。
```bash
wget https://github.com/ユーザー名/リポジトリ名/releases/download/v1.0.0/pve-wol
```

### Python HTTPサーバーを使用する（開発中の一時転送用）
開発機でビルドしたバイナリをすぐに転送したい場合は、開発機のターミナルで簡易サーバーを立てます。
```bash
make build
python3 -m http.server 8000
```
LXCコンソールから開発機のIPを指定してダウンロードします。
```bash
wget http://開発機のIPアドレス:8000/pve-wol
```

## デプロイ方法 (OpenRC)
LXC (Alpine Linux) 上でのサービス化と要塞化の手順です。

### 1. 鍵の生成と配置, バイナリの配置
LXC内でSSH鍵を生成し、対象のWindows PCの `authorized_keys` に登録します。
```bash
mkdir -p /opt/pve-wol
ssh-keygen -t ed25519 -f /opt/pve-wol/id_ed25519 -N ""
```
※Windows管理者の場合、登録先は `C:\ProgramData\ssh\administrators_authorized_keys` になります。権限設定にご注意ください。

### 2. 環境変数の設定
`/etc/conf.d/pve-wol` を作成し、設定値を記述します。OpenRCから環境変数としてGoプログラムに引き継がせるため、必ず行頭に `export` を付与してください。
```bash
export LISTEN_ADDR="10.0.0.X:8080"
export BCAST_ADDR="192.168.40.255"
export TARGET_MAC="XX:XX:XX:XX:XX:XX"
export TARGET_IP="192.168.40.51:22"
export TARGET_USER="John Doe"
export SSH_KEY_PATH="/opt/pve-wol/id_ed25519"
```

### 3. OpenRCスクリプトの作成
`/etc/init.d/pve-wol` を作成します。
```bash
#!/sbin/openrc-run

name="pve-wol"
command="/opt/pve-wol/pve-wol"
command_background="yes"
pidfile="/run/${name}.pid"

depend() {
    need net
}
```

### 4. サービスの起動と要塞化
権限を付与し、自動起動を有効化します。動作確認後、SSHデーモンを削除します。
```bash
chmod +x /opt/pve-wol/pve-wol
chmod +x /etc/init.d/pve-wol
rc-update add pve-wol default
rc-service pve-wol start

# 要塞化（SSHデーモンの完全削除）
rc-service sshd stop
rc-update del sshd default
apk del openssh
```

## 開発環境立ち上げ
ローカル開発機での立ち上げ手順です。

### 1. 環境変数の準備
```bash
cp .env.example .env
# .env を開いてご自身の環境に合わせて修正してください
```

### 2. 特権なしPingの許可（Linuxのみ）
ローカル開発機で `GET /status` (ICMP Ping) をテストするために、Pingパケットの生成を許可します。
```bash
make setup-ping
```

### 3. サーバーの起動
以下のコマンドでビルドと起動が行われます。`.env` は自動的に読み込まれます。
```bash
make run
```