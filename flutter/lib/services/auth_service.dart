import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'api_service.dart';
import '../models/user.dart';

class AuthService extends ChangeNotifier {
  String? _token;
  User? _user;

  String? get token => _token;
  User? get user => _user;
  bool get isAuthenticated => _token != null;

  Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    _token = prefs.getString('token');
    final userJson = prefs.getString('user');
    if (userJson != null) {
      try {
        _user = User.fromJson(jsonDecode(userJson));
      } catch (_) {
        _token = null;
      }
    }
    notifyListeners();
  }

  Future<void> login(String email, String password) async {
    final response = await ApiService.request('POST', '/api/auth/login',
        body: {'email': email, 'password': password});
    _token = response['data']['token'];
    _user = User.fromJson(response['data']['user']);

    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('token', _token!);
    await prefs.setString('user', jsonEncode(response['data']['user']));
    notifyListeners();
  }

  Future<void> logout() async {
    if (_token != null) {
      try {
        await ApiService.request('POST', '/api/auth/logout', token: _token);
      } catch (_) {}
    }
    _token = null;
    _user = null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('token');
    await prefs.remove('user');
    notifyListeners();
  }
}
