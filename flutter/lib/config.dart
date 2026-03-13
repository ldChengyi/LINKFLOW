class AppConfig {
  static const String appName = 'LinkFlow';
  static const int defaultPort = 8080;
  static const Duration heartbeatInterval = Duration(minutes: 2);
  static const Duration requestTimeout = Duration(seconds: 15);
  static const Duration voiceTimeout = Duration(seconds: 30);
}
