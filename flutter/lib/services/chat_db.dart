import 'package:sqflite/sqflite.dart';
import 'package:path/path.dart';
import '../models/chat_message.dart';

class ChatDB {
  static Database? _db;

  static Future<void> init() async {
    if (_db != null) return;
    final dbPath = await getDatabasesPath();
    _db = await openDatabase(
      join(dbPath, 'chat.db'),
      version: 1,
      onCreate: (db, version) async {
        await db.execute('''
          CREATE TABLE messages (
            id TEXT PRIMARY KEY,
            device_id TEXT NOT NULL,
            type TEXT NOT NULL,
            content TEXT NOT NULL,
            audio_url TEXT,
            action TEXT,
            success INTEGER,
            data TEXT,
            timestamp INTEGER NOT NULL
          )
        ''');
        await db.execute(
          'CREATE INDEX idx_messages_device ON messages(device_id, timestamp DESC)',
        );
      },
    );
  }

  static Future<void> insertMessage(String deviceId, ChatMessage msg) async {
    await _db?.insert('messages', msg.toMap(deviceId),
        conflictAlgorithm: ConflictAlgorithm.replace);
  }

  static Future<List<ChatMessage>> getMessages(String deviceId,
      {int limit = 50}) async {
    if (_db == null) return [];
    final rows = await _db!.query(
      'messages',
      where: 'device_id = ?',
      whereArgs: [deviceId],
      orderBy: 'timestamp DESC',
      limit: limit,
    );
    return rows.map((r) => ChatMessage.fromMap(r)).toList();
  }

  static Future<void> deleteDeviceMessages(String deviceId) async {
    await _db?.delete('messages',
        where: 'device_id = ?', whereArgs: [deviceId]);
  }
}
