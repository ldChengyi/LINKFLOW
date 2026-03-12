// =============================================================================
// LinkFlow ESP32 WS2812 LED Strip Module — Adafruit NeoPixel Implementation
// =============================================================================

#include "ws2812.h"

bool WS2812Module::begin() {
    _strip.begin();
    _strip.setBrightness(_brightness);
    _strip.clear();
    _strip.show();
    Serial.println(F("[WS2812] LED strip ready"));
    return true;
}

void WS2812Module::update() {
    if (!_on) return;
    if (_color != COLOR_RAINBOW) return;

    // 彩虹动画：每 30ms 推进一步
    unsigned long now = millis();
    if (now - _lastRainbow < 30) return;
    _lastRainbow = now;

    applyRainbow();
    _rainbowOffset += 256;
}

void WS2812Module::setSwitch(bool on) {
    _on = on;
    if (_on) {
        _strip.setBrightness(_brightness);
        applyColor();
    } else {
        _strip.clear();
        _strip.show();
    }
    Serial.printf("[WS2812] switch = %s\n", _on ? "ON" : "OFF");
}

void WS2812Module::setColor(uint8_t colorEnum) {
    if (colorEnum > COLOR_RAINBOW) colorEnum = COLOR_RED;
    _color = colorEnum;
    if (_on) applyColor();
    Serial.printf("[WS2812] color = %d\n", _color);
}

void WS2812Module::setBrightness(uint8_t brightness) {
    _brightness = brightness;
    _strip.setBrightness(_brightness);
    if (_on) _strip.show();
    Serial.printf("[WS2812] brightness = %d\n", _brightness);
}

uint32_t WS2812Module::enumToColor(uint8_t c) {
    switch (c) {
        case COLOR_RED:    return _strip.Color(255, 0, 0);
        case COLOR_GREEN:  return _strip.Color(0, 255, 0);
        case COLOR_BLUE:   return _strip.Color(0, 0, 255);
        case COLOR_WHITE:  return _strip.Color(255, 255, 255);
        case COLOR_YELLOW: return _strip.Color(255, 255, 0);
        case COLOR_PURPLE: return _strip.Color(128, 0, 255);
        case COLOR_CYAN:   return _strip.Color(0, 255, 255);
        default:           return _strip.Color(255, 0, 0);
    }
}

void WS2812Module::applyColor() {
    if (_color == COLOR_RAINBOW) {
        applyRainbow();
    } else {
        uint32_t c = enumToColor(_color);
        _strip.fill(c);
        _strip.show();
    }
}

void WS2812Module::applyRainbow() {
    for (int i = 0; i < _strip.numPixels(); i++) {
        uint16_t hue = _rainbowOffset + (i * 65536 / _strip.numPixels());
        _strip.setPixelColor(i, _strip.gamma32(_strip.ColorHSV(hue)));
    }
    _strip.show();
}
