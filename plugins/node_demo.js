#!/usr/bin/env node
// @name: node_demo
// @desc: Node.js / JavaScript yerel eklenti çalışma zamanı gösterimi
// @category: RUNTIME
// @perm: net, stream, eventbus
// @tags: nodejs, javascript, dynamic, v8

const path = require('path');
const { MT, Color, Style } = require(path.join(__dirname, 'sdk', 'mt.js'));

function main() {
  const mt = new MT();

  console.log(Style.badge('NODE.JS RUNTIME V8', Color.DARK, Color.EMERALD));
  console.log(Style.color('MT Core JavaScript Plugin Host Initialized\n', Color.MUTED));

  // 1. Bilgi Kartı
  const card = Style.card('NODE PROCESS INFO', {
    'Node Sürümü': process.version,
    'Platform': process.platform,
    'PID': process.pid,
    'Uptime': `${Math.floor(process.uptime())}s`,
    'Bellek Kullanımı': `${Math.round(process.memoryUsage().heapUsed / 1024 / 1024)} MB`
  }, Color.EMERALD, 48);
  console.log(card);
  console.log();

  // 2. Tablo Render İsteği
  mt.renderTable(
    ['Paket', 'Tür', 'Versiyon', 'Bağımlılık'],
    [
      ['@google/genai', 'SDK', '0.1.2', 'Node 18+'],
      ['express', 'Server', '4.19.2', 'Node 14+'],
      ['ws', 'WebSocket', '8.16.0', 'Node 12+'],
      ['zod', 'Validation', '3.22.4', 'Zero-dep']
    ]
  );

  // 3. Dinamik Komut Enjeksiyonu
  mt.registerCommand(
    'v8info',
    'V8 ENGINE STATS',
    'NODEJS',
    'Node.js V8 heap istatistiklerini görüntüler',
    'node -e "console.log(JSON.stringify(v8.getHeapStatistics(), null, 2))"'
  );

  // 4. Toast & Event Bus
  mt.toast('Node.js eklentisi başarıyla çalıştırıldı!');
  mt.emitEvent('node_ready', { version: process.version, pid: process.pid });
  mt.kvSet('NODE_LAST_RUN', new Date().toISOString());
}

main();
