.PHONY: all build run clean tidy setup-ping

# バイナリ名
BINARY_NAME=pve-wol
# ソースディレクトリ
SRC_DIR=./src

all: build

# LXC環境(Alpine Linux)向けにバイナリをビルドする
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) $(SRC_DIR)

# 開発用：手元でサーバーを起動する（.env が自動で読み込まれます）
run:
	go run $(SRC_DIR)

# 依存関係の整理とコードのフォーマットを行う
tidy:
	go fmt $(SRC_DIR)/...
	go mod tidy

# 生成されたバイナリを削除する
clean:
	go clean
	rm -f $(BINARY_NAME)

# 開発用：現在のユーザー向けに特権なしPing(Unprivileged Ping)を一時的に許可する
setup-ping:
	@echo "現在のユーザー(GID: $$(id -g))に特権なしPingを許可します。sudoパスワードを求められる場合があります。"
	sudo sysctl -w net.ipv4.ping_group_range="$$(id -g) $$(id -g)"
