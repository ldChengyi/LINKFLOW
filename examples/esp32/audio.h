// =============================================================================
// LinkFlow ESP32 Audio Module — ES8311 + NS4150B
// =============================================================================

#ifndef AUDIO_H
#define AUDIO_H

#include <Arduino.h>
#include <driver/i2s.h>
#include <Wire.h>

// I2S 引脚
#define I2S_BCLK 26
#define I2S_LRCK 25
#define I2S_DOUT 22
// #define I2S_MCLK 0   // 如果板子有 GPIO0 引出，取消注释

// I2C 引脚（ES8311 配置）
#define I2C_SCL 27
#define I2C_SDA 14

// 功放使能（如果模块有 SD/EN 引脚，取消注释下面这行并连接到 GPIO13）
// #define AMP_SD 13

// ES8311 I2C 地址
#define ES8311_ADDR 0x18

class AudioModule {
public:
    bool begin();
    bool ready() const { return _ready; }
    void enable();
    void disable();
    void deinit();   // 卸载 I2S 驱动（TTS 播放前调用）
    void reinit();   // 重装 I2S 驱动（TTS 播放后调用）
    bool playTone(uint16_t freq, uint16_t duration_ms);

    // 反馈音效
    void playBeep();        // 属性变更确认：1000Hz 短鸣 80ms
    void playSuccessTone(); // 语音成功：800Hz→1200Hz 升调
    void playErrorTone();   // 语音失败：600Hz→300Hz 降调

private:
    bool _ready = false;
    bool initES8311();
    void writeES8311(uint8_t reg, uint8_t val);
    uint8_t readES8311(uint8_t reg);
};

#endif
