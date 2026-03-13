import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import '../config.dart';

class ApiService {
  static String? _baseUrl;

  static Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    final serverUrl = prefs.getString('serverUrl');
    if (serverUrl != null) _baseUrl = serverUrl;
  }

  static Future<void> setServerUrl(String url) async {
    _baseUrl = url;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('serverUrl', url);
  }

  static String? get baseUrl => _baseUrl;

  static Future<Map<String, dynamic>> request(
    String method,
    String path, {
    Map<String, dynamic>? body,
    String? token,
    Duration? timeout,
  }) async {
    if (_baseUrl == null) throw Exception('Server URL not configured');

    final uri = Uri.parse('$_baseUrl$path');
    final headers = {'Content-Type': 'application/json'};
    if (token != null) headers['Authorization'] = 'Bearer $token';
    final dur = timeout ?? AppConfig.requestTimeout;

    http.Response response;
    try {
      if (method == 'GET') {
        response = await http.get(uri, headers: headers).timeout(dur);
      } else if (method == 'POST') {
        response = await http.post(uri, headers: headers, body: jsonEncode(body)).timeout(dur);
      } else if (method == 'PUT') {
        response = await http.put(uri, headers: headers, body: jsonEncode(body)).timeout(dur);
      } else {
        throw Exception('Unsupported method: $method');
      }
    } catch (e) {
      throw Exception('Network error: $e');
    }

    final data = jsonDecode(response.body);
    if (data['code'] != 200) throw Exception(data['msg'] ?? 'Request failed');
    return data;
  }
}
