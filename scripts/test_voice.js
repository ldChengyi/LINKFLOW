const mqtt = require('mqtt');

const DEVICE_ID = '528d85fb-1421-4e65-b7b7-19a3cc2dfa3b';
const DEVICE_SECRET = '4a05f619e5a0307b9ed4de8d190b4b18bf9ffce02c9a453c81031681a5ebd1d1';

const client = mqtt.connect('mqtt://localhost:1883', {
  clientId: DEVICE_ID,
  username: DEVICE_ID,
  password: DEVICE_SECRET,
});

const tests = [
  { text: '打开客厅灯', delay: 1000 },
  { text: '把客厅灯亮度调到80', delay: 4000 },
  { text: '关闭客厅灯', delay: 7000 },
];

client.on('connect', () => {
  console.log('MQTT connected as', DEVICE_ID);

  // 订阅回复和下发 topic
  client.subscribe([
    `devices/${DEVICE_ID}/voice/down`,
    `devices/${DEVICE_ID}/telemetry/down`,
    `devices/${DEVICE_ID}/service/invoke`,
  ]);

  // 依次发送测试指令
  tests.forEach(({ text, delay }) => {
    setTimeout(() => {
      console.log(`\n>>> 发送语音指令: "${text}"`);
      client.publish(
        `devices/${DEVICE_ID}/voice/up`,
        JSON.stringify({ text }),
        { qos: 1 }
      );
    }, delay);
  });

  // 10秒后退出
  setTimeout(() => {
    console.log('\n=== 测试完成 ===');
    client.end();
    process.exit(0);
  }, 12000);
});

client.on('message', (topic, message) => {
  const suffix = topic.split('/').slice(2).join('/');
  console.log(`<<< [${suffix}] ${message.toString()}`);
});

client.on('error', (err) => {
  console.error('MQTT error:', err.message);
  process.exit(1);
});
