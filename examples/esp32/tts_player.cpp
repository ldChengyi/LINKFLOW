// =============================================================================
// LinkFlow ESP32 TTS Player — MP3 HTTP Streaming Playback Implementation
// =============================================================================
// 使用 ESP8266Audio 库流式解码 MP3 并通过 I2S 输出到 ES8311
// 播放前卸载 AudioModule 的 I2S 驱动，播放后恢复（避免冲突）
// =============================================================================

#include "tts_player.h"
#include "audio.h"

#include <AudioGeneratorMP3.h>
#include <AudioOutputI2S.h>
#include <AudioFileSourceHTTPStream.h>
#include <AudioFileSourceBuffer.h>

void TTSPlayer::begin(AudioModule *audio)
{
    _audio = audio;
    Serial.println(F("[TTS] Player ready"));
}

bool TTSPlayer::play(const char *url)
{
    if (!url || strlen(url) == 0)
    {
        Serial.println(F("[TTS] Empty URL, skipping"));
        return false;
    }

    Serial.printf("[TTS] Streaming: %s\n", url);
    Serial.printf("[TTS] Free heap before play: %d bytes\n", ESP.getFreeHeap());

    // 1. 卸载现有 I2S 驱动（AudioModule 占用的）
    if (_audio)
    {
        _audio->enable(); // 确保功放开启
        _audio->deinit();
    }

    // 2. 创建 ESP8266Audio 播放链
    //    HTTPStream → Buffer(4KB) → MP3 Decoder → I2S Output
    AudioFileSourceHTTPStream *httpSrc = new AudioFileSourceHTTPStream(url);
    if (!httpSrc)
    {
        Serial.println(F("[TTS] Failed to create HTTP source"));
        if (_audio) _audio->reinit();
        return false;
    }

    AudioFileSourceBuffer *buffer = new AudioFileSourceBuffer(httpSrc, 4096);
    AudioOutputI2S *i2sOut = new AudioOutputI2S(0); // I2S_NUM_0
    i2sOut->SetPinout(I2S_BCLK, I2S_LRCK, I2S_DOUT);
    i2sOut->SetGain(1.0);

    AudioGeneratorMP3 *mp3 = new AudioGeneratorMP3();

    bool success = false;
    if (mp3->begin(buffer, i2sOut))
    {
        Serial.println(F("[TTS] Playing..."));
        unsigned long startTime = millis();

        while (mp3->isRunning())
        {
            if (!mp3->loop()) break;

            // 超时保护：最长 30 秒
            if (millis() - startTime > 30000)
            {
                Serial.println(F("[TTS] Playback timeout (30s)"));
                break;
            }
            yield(); // 喂狗，防止 WDT 重启
        }

        success = true;
        Serial.printf("[TTS] Playback done, duration=%lums\n", millis() - startTime);
    }
    else
    {
        Serial.println(F("[TTS] MP3 begin failed — check URL or audio format"));
    }

    // 3. 清理
    mp3->stop();
    delete mp3;
    delete buffer;
    delete httpSrc;
    delete i2sOut;

    // 4. 恢复 AudioModule 的 I2S 驱动（用于提示音）
    if (_audio) _audio->reinit();

    Serial.printf("[TTS] Free heap after play: %d bytes\n", ESP.getFreeHeap());
    return success;
}
