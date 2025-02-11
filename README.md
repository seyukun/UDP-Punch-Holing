# UDP Punch Holing
ポート解放を不要としたP2P式の簡易的なメッセージアプリです

# Usage
Serverはパンチホーリングで開いたIP:Portを管理するためにあります  
(メッセージを受け取るものではない)
```bash
# Server
yarn # install
yarn tsc # build
node dest/main.js # run
```

ClientはServerから取得したアドレスリストに接続し、
メッセージを送ります（readlineしているので任意の文がマルチキャストで送れます）
```bash
# Client
go build .
./stun-punch-holing http://server:3000
```
