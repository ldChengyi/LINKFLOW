// =============================================================================
// LinkFlow ESP32 WS2812 LED Strip Module
// =============================================================================

#ifndef WS2812_H
#define WS2812_H

#include <Arduino.h>
#include <Adafruit_NeoPixel.h>

// WS2812 配置
#define WS2812_PIN 15
#define WS2812_NUM_LEDS 15 // 灯珠数量，按实际修改

// 颜色枚举（与物模型 led_color enum 对应）
enum LedColor : uint8_t
{
    COLOR_RED = 0,
    COLOR_GREEN = 1,
    COLOR_BLUE = 2,
    COLOR_WHITE = 3,
    COLOR_YELLOW = 4,
    COLOR_PURPLE = 5,
    COLOR_CYAN = 6,
    COLOR_RAINBOW = 7
};

class WS2812Module
{
public:
    bool begin();
    void update(); // loop 中调用，驱动彩虹动画

    void setSwitch(bool on);
    void setColor(uint8_t colorEnum);
    void setBrightness(uint8_t brightness);

    bool isOn() const { return _on; }
    uint8_t getColor() const { return _color; }
    uint8_t getBrightness() const { return _brightness; }

private:
    Adafruit_NeoPixel _strip{WS2812_NUM_LEDS, WS2812_PIN, NEO_GRB + NEO_KHZ800};
    bool _on = false;
    uint8_t _color = COLOR_RED;
    uint8_t _brightness = 128;
    uint16_t _rainbowOffset = 0;
    unsigned long _lastRainbow = 0;

    uint32_t enumToColor(uint8_t c);
    void applyColor();
    void applyRainbow();
};

#endif
