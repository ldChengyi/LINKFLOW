import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'services/api_service.dart';
import 'services/auth_service.dart';
import 'services/chat_db.dart';
import 'services/ws_service.dart';
import 'screens/server_config_screen.dart';
import 'screens/login_screen.dart';
import 'screens/home_screen.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await ApiService.init();
  await ChatDB.init();
  final authService = AuthService();
  await authService.init();
  final wsService = WSService();

  // 如果已登录，立即连接 WS
  if (authService.isAuthenticated && ApiService.baseUrl != null) {
    wsService.connect(ApiService.baseUrl!, authService.token!);
  }

  // 监听登录/登出，自动连接/断开 WS
  authService.addListener(() {
    if (authService.isAuthenticated && ApiService.baseUrl != null) {
      wsService.connect(ApiService.baseUrl!, authService.token!);
    } else {
      wsService.disconnect();
    }
  });

  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider.value(value: authService),
        ChangeNotifierProvider.value(value: wsService),
      ],
      child: const MyApp(),
    ),
  );
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'LinkFlow',
      theme: ThemeData(
        primaryColor: const Color(0xFF2196F3),
        scaffoldBackgroundColor: Colors.white,
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF2196F3),
          brightness: Brightness.light,
        ),
        appBarTheme: const AppBarTheme(
          backgroundColor: Color(0xFF2196F3),
          foregroundColor: Colors.white,
          elevation: 0,
        ),
      ),
      home: const SplashScreen(),
      routes: {
        '/server-config': (_) => const ServerConfigScreen(),
        '/login': (_) => const LoginScreen(),
        '/home': (_) => const HomeScreen(),
      },
    );
  }
}

class SplashScreen extends StatefulWidget {
  const SplashScreen({super.key});

  @override
  State<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends State<SplashScreen> {
  @override
  void initState() {
    super.initState();
    _navigate();
  }

  Future<void> _navigate() async {
    await Future.delayed(const Duration(milliseconds: 500));
    if (!mounted) return;

    final hasServer = ApiService.baseUrl != null;
    final isAuth = context.read<AuthService>().isAuthenticated;

    if (!hasServer) {
      Navigator.pushReplacementNamed(context, '/server-config');
    } else if (!isAuth) {
      Navigator.pushReplacementNamed(context, '/login');
    } else {
      Navigator.pushReplacementNamed(context, '/home');
    }
  }

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      backgroundColor: Colors.white,
      body: Center(child: CircularProgressIndicator(color: Color(0xFF2196F3))),
    );
  }
}
