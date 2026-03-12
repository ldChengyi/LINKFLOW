// =============================================================================
// LinkFlow ESP32 Audio Module — ES8311 + NS4150B Implementation
// =============================================================================

#include "audio.h"

bool AudioModule::begin()
{
    Serial.println(F("[Audio] ---- Init Start ----"));

    // 初始化功放使能引脚（如果定义了）
#ifdef AMP_SD
    pinMode(AMP_SD, OUTPUT);
    digitalWrite(AMP_SD, LOW);
    Serial.printf("[Audio] AMP_SD pin %d configured, set LOW\n", AMP_SD);
#else
    Serial.println(F("[Audio] AMP_SD not defined, skipping amp enable pin"));
#endif

    // 初始化 I2C
    Serial.printf("[Audio] I2C begin: SDA=%d, SCL=%d\n", I2C_SDA, I2C_SCL);
    if (!Wire.begin(I2C_SDA, I2C_SCL))
    {
        Serial.println(F("[Audio] I2C bus init FAILED"));
        return false;
    }
    Wire.setClock(100000);
    Serial.println(F("[Audio] I2C bus ready, clock=100kHz"));

    // I2C 扫描：列出总线上所有设备
    Serial.println(F("[Audio] I2C scanning..."));
    int found = 0;
    for (uint8_t addr = 1; addr < 127; addr++)
    {
        Wire.beginTransmission(addr);
        if (Wire.endTransmission() == 0)
        {
            Serial.printf("[Audio] I2C device found at 0x%02X\n", addr);
            found++;
        }
    }
    if (found == 0)
    {
        Serial.println(F("[Audio] I2C scan: NO devices found! Check SDA/SCL wiring and module power"));
        return false;
    }
    Serial.printf("[Audio] I2C scan: %d device(s) found\n", found);

    // 初始化 ES8311
    Serial.printf("[Audio] ES8311 init at address 0x%02X...\n", ES8311_ADDR);
    if (!initES8311())
    {
        Serial.println(F("[Audio] ES8311 init FAILED — device not responding at expected address"));
        return false;
    }
    Serial.println(F("[Audio] ES8311 init OK"));

    // 初始化 I2S
    Serial.printf("[Audio] I2S config: BCLK=%d, LRCK=%d, DOUT=%d, rate=16000, 16bit mono\n",
                  I2S_BCLK, I2S_LRCK, I2S_DOUT);

    i2s_config_t i2s_config = {
        .mode = (i2s_mode_t)(I2S_MODE_MASTER | I2S_MODE_TX),
        .sample_rate = 16000,
        .bits_per_sample = I2S_BITS_PER_SAMPLE_16BIT,
        .channel_format = I2S_CHANNEL_FMT_ONLY_LEFT,
        .communication_format = I2S_COMM_FORMAT_STAND_I2S,
        .intr_alloc_flags = ESP_INTR_FLAG_LEVEL1,
        .dma_buf_count = 8,
        .dma_buf_len = 64,
        .use_apll = false,
        .tx_desc_auto_clear = true,
        .fixed_mclk = 0};

    i2s_pin_config_t pin_config = {
        .bck_io_num = I2S_BCLK,
        .ws_io_num = I2S_LRCK,
        .data_out_num = I2S_DOUT,
        .data_in_num = I2S_PIN_NO_CHANGE};

    esp_err_t err = i2s_driver_install(I2S_NUM_0, &i2s_config, 0, NULL);
    if (err != ESP_OK)
    {
        Serial.printf("[Audio] I2S driver install FAILED, err=0x%X\n", err);
        return false;
    }
    Serial.println(F("[Audio] I2S driver installed"));

    err = i2s_set_pin(I2S_NUM_0, &pin_config);
    if (err != ESP_OK)
    {
        Serial.printf("[Audio] I2S set pin FAILED, err=0x%X\n", err);
        return false;
    }
    Serial.println(F("[Audio] I2S pins configured"));

    Serial.println(F("[Audio] ---- Init OK ----"));
    _ready = true;
    return true;
}

void AudioModule::enable()
{
#ifdef AMP_SD
    digitalWrite(AMP_SD, HIGH);
    delay(10);
#endif
}

void AudioModule::disable()
{
#ifdef AMP_SD
    digitalWrite(AMP_SD, LOW);
#endif
}

bool AudioModule::playTone(uint16_t freq, uint16_t duration_ms)
{
    if (!_ready)
    {
        Serial.println(F("[Audio] playTone skipped — not ready"));
        return false;
    }

    const int sample_rate = 16000;
    const int samples = (sample_rate * duration_ms) / 1000;
    size_t bufSize = samples * sizeof(int16_t);
    int16_t *buffer = (int16_t *)malloc(bufSize);

    if (!buffer)
    {
        Serial.printf("[Audio] playTone malloc FAILED, need %d bytes, free heap=%d\n",
                      bufSize, ESP.getFreeHeap());
        return false;
    }

    // 生成正弦波
    for (int i = 0; i < samples; i++)
    {
        float t = (float)i / sample_rate;
        buffer[i] = (int16_t)(sin(2 * PI * freq * t) * 32000);
    }

    enable();
    size_t written;
    esp_err_t err = i2s_write(I2S_NUM_0, buffer, bufSize, &written, portMAX_DELAY);

    if (err != ESP_OK)
    {
        Serial.printf("[Audio] i2s_write FAILED, err=0x%X\n", err);
    }
    else if (written != bufSize)
    {
        Serial.printf("[Audio] i2s_write partial: %d/%d bytes\n", written, bufSize);
    }

    free(buffer);
    return (err == ESP_OK);
}

bool AudioModule::initES8311()
{
    // 检测 ES8311
    Wire.beginTransmission(ES8311_ADDR);
    uint8_t ack = Wire.endTransmission();
    if (ack != 0)
    {
        Serial.printf("[Audio] ES8311 ACK failed at 0x%02X, error=%d", ES8311_ADDR, ack);
        switch (ack)
        {
        case 1:
            Serial.println(F(" (data too long)"));
            break;
        case 2:
            Serial.println(F(" (NACK on address — device not found)"));
            break;
        case 3:
            Serial.println(F(" (NACK on data)"));
            break;
        case 4:
            Serial.println(F(" (other error)"));
            break;
        case 5:
            Serial.println(F(" (timeout)"));
            break;
        default:
            Serial.println();
            break;
        }
        return false;
    }
    Serial.println(F("[Audio] ES8311 ACK OK"));

    // 完整 DAC 配置（BCLK 作为内部时钟源，不需要外部 MCLK）
    Serial.println(F("[Audio] ES8311 writing registers..."));

    // 1. 软复位
    writeES8311(0x00, 0x1F);
    delay(20);
    writeES8311(0x00, 0x00); // 解除复位
    delay(10);
    Serial.println(F("[Audio]   Reset done"));

    // 2. 时钟配置 — 使用 BCLK 作为时钟源（无需 MCLK）
    writeES8311(0x01, 0x3F); // CLK Manager: BCLK as clock source, auto-detect
    Serial.println(F("[Audio]   REG 0x01=0x3F (BCLK clock source)"));
    writeES8311(0x02, 0x00); // CLK DIV1
    Serial.println(F("[Audio]   REG 0x02=0x00 (CLK DIV1)"));
    writeES8311(0x03, 0x10); // CLK DIV2
    Serial.println(F("[Audio]   REG 0x03=0x10 (CLK DIV2)"));
    writeES8311(0x04, 0x10); // CLK ADC/DAC divider
    Serial.println(F("[Audio]   REG 0x04=0x10 (ADC/DAC div)"));
    writeES8311(0x05, 0x00); // CLK Manager 5
    Serial.println(F("[Audio]   REG 0x05=0x00 (CLK5)"));

    // 3. I2S 格式配置
    writeES8311(0x09, 0x0C); // SDP In: 16-bit, I2S format
    Serial.println(F("[Audio]   REG 0x09=0x0C (I2S 16bit in)"));
    writeES8311(0x0A, 0x0C); // SDP Out: 16-bit, I2S format
    Serial.println(F("[Audio]   REG 0x0A=0x0C (I2S 16bit out)"));

    // 4. DAC 系统配置
    writeES8311(0x0D, 0x01); // System: power up
    Serial.println(F("[Audio]   REG 0x0D=0x01 (System power up)"));
    writeES8311(0x0E, 0x02); // System: DAC on
    Serial.println(F("[Audio]   REG 0x0E=0x02 (DAC on)"));
    writeES8311(0x12, 0x00); // System: normal operation
    Serial.println(F("[Audio]   REG 0x12=0x00 (Normal mode)"));

    // 5. DAC 音量
    writeES8311(0x32, 0xFF); // DAC volume (最大)
    Serial.println(F("[Audio]   REG 0x32=0xFF (DAC volume MAX)"));

    // 6. 模拟输出配置
    writeES8311(0x13, 0x10); // ADC/DAC power on
    Serial.println(F("[Audio]   REG 0x13=0x10 (DAC power on)"));
    writeES8311(0x14, 0x1A); // Lineout power + DAC to output mixer
    Serial.println(F("[Audio]   REG 0x14=0x1A (Lineout + mixer on)"));
    writeES8311(0x37, 0x08); // Mixer: DAC output to analog out
    Serial.println(F("[Audio]   REG 0x37=0x08 (Mixer DAC→out)"));
    writeES8311(0x44, 0x08); // Analog output gain (max)
    Serial.println(F("[Audio]   REG 0x44=0x08 (Analog gain max)"));

    delay(50); // 等待稳定

    // 回读验证
    uint8_t val = readES8311(0x01);
    Serial.printf("[Audio] ES8311 readback REG 0x01 = 0x%02X (expect 0x3F)\n", val);

    return true;
}

void AudioModule::writeES8311(uint8_t reg, uint8_t val)
{
    Wire.beginTransmission(ES8311_ADDR);
    Wire.write(reg);
    Wire.write(val);
    Wire.endTransmission();
}

uint8_t AudioModule::readES8311(uint8_t reg)
{
    Wire.beginTransmission(ES8311_ADDR);
    Wire.write(reg);
    Wire.endTransmission(false);
    Wire.requestFrom(ES8311_ADDR, 1);
    return Wire.read();
}

// ======================== I2S 生命周期（TTS 播放需要） ========================

void AudioModule::deinit()
{
    if (!_ready) return;
    i2s_driver_uninstall(I2S_NUM_0);
    _ready = false;
    Serial.println(F("[Audio] I2S driver uninstalled (for TTS)"));
}

void AudioModule::reinit()
{
    if (_ready) return;

    i2s_config_t i2s_config = {
        .mode = (i2s_mode_t)(I2S_MODE_MASTER | I2S_MODE_TX),
        .sample_rate = 16000,
        .bits_per_sample = I2S_BITS_PER_SAMPLE_16BIT,
        .channel_format = I2S_CHANNEL_FMT_ONLY_LEFT,
        .communication_format = I2S_COMM_FORMAT_STAND_I2S,
        .intr_alloc_flags = ESP_INTR_FLAG_LEVEL1,
        .dma_buf_count = 8,
        .dma_buf_len = 64,
        .use_apll = false,
        .tx_desc_auto_clear = true,
        .fixed_mclk = 0};

    i2s_pin_config_t pin_config = {
        .bck_io_num = I2S_BCLK,
        .ws_io_num = I2S_LRCK,
        .data_out_num = I2S_DOUT,
        .data_in_num = I2S_PIN_NO_CHANGE};

    if (i2s_driver_install(I2S_NUM_0, &i2s_config, 0, NULL) != ESP_OK)
    {
        Serial.println(F("[Audio] I2S reinit FAILED"));
        return;
    }
    if (i2s_set_pin(I2S_NUM_0, &pin_config) != ESP_OK)
    {
        Serial.println(F("[Audio] I2S reinit pin FAILED"));
        return;
    }
    _ready = true;
    Serial.println(F("[Audio] I2S driver reinstalled"));
}

// ======================== 反馈音效 ========================

void AudioModule::playBeep()
{
    // 属性变更确认：1000Hz 短鸣 80ms
    enable();
    playTone(1000, 80);
    disable();
}

void AudioModule::playSuccessTone()
{
    // 语音成功：800Hz→1200Hz 升调（两段）
    enable();
    playTone(800, 120);
    delay(40);
    playTone(1200, 120);
    disable();
}

void AudioModule::playErrorTone()
{
    // 语音失败：600Hz→300Hz 降调（两段）
    enable();
    playTone(600, 160);
    delay(60);
    playTone(300, 160);
    disable();
}
