// =============================================================================
// LinkFlow ESP32 TTS Player Module — MP3 HTTP Stream + ES8311 Playback
// =============================================================================
// 流程：收到 voice/down 中的 audio_url → HTTP 流式下载 MP3 → 解码 → I2S 播放
// 依赖库：ESP8266Audio（PlatformIO: earlephilhower/ESP8266Audio）
// =============================================================================

#ifndef TTS_PLAYER_H
#define TTS_PLAYER_H

#include <Arduino.h>

class AudioModule; // forward declare

class TTSPlayer
{
public:
    void begin(AudioModule *audio);
    bool play(const char *url); // 阻塞式：下载 + 解码 + 播放，完成后返回

private:
    AudioModule *_audio = nullptr;
};

#endif
