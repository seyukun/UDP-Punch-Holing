import express from "express";
import Keyv from "Keyv";

// Keyv の初期化（ここでは SQLite を利用）
// SQLite を利用する場合、接続文字列例: "sqlite://./database.sqlite"
// メモリ上での利用の場合は、引数を省略または "" を指定できます。
const keyv = new Keyv();

// リクエストボディの型定義
interface Register {
  address: string;
  sessionId: string;
}

const app = express();
const PORT = process.env.PORT || 3000;

// JSON ボディのパースを有効にする
app.use(express.json());

/**
 * GET "/"
 * Keyv からピアアドレス一覧を取得し、以下の形式でレスポンスを返します:
 * { "peers": [ "アドレス1", "アドレス2", ... ] }
 */
app.get("/", async (_, res) => {
  try {
    const peerKeys = ((await keyv.get("0")) as string[]) || [];
    const peersPromise = peerKeys.map(async (peerKey) => {
      return (await keyv.get(peerKey)) as Register;
    });
    const peers = await Promise.all(peersPromise);
    const addresses = peers.map((peer) => peer.address);
    console.log("Peers:", addresses);
    res.end(JSON.stringify({ peers: addresses }));
  } catch (error) {
    console.error("Error retrieving peers:", error);
    res.status(500).json({ error: (error as Error).toString() });
  }
});

/**
 * POST "/"
 * リクエストボディ (JSON) の形式:
 * { "address": "ピアのアドレス文字列", "sessionId": "セッションID文字列" }
 *
 * 未登録の場合のみ新規にピアアドレスを Keyv に保存します。
 */
app.post("/", async (req, res) => {
  const { address, sessionId } = req.body as Register;
  if (!address || !sessionId) {
    res.status(400).json({ error: "require address,sessionId" });
    return;
  }

  try {
    let peerKeys = ((await keyv.get("0")) as string[]) || [];
    if (!peerKeys.includes(sessionId)) {
      for (const peerKey of peerKeys) {
        const peer = (await keyv.get(peerKey)) as Register | null;
        if (peer == null) {
          peerKeys = peerKeys.filter((key) => key != peerKey);
        }
      }
      peerKeys.push(sessionId);
      await keyv.set("0", peerKeys, 1000 * 25);
    }
    await keyv.set(sessionId, req.body as Register, 1000 * 25);
    res.json({ status: "ok" });
  } catch (error) {
    console.error("Error in POST /:", error);
    res.status(500).json({ error: (error as Error).toString() });
  }
});

// サーバー起動
app.listen(PORT, () => {
  console.log(`Server Started on ${PORT}`);
});
