import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器 - 添加 token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// 响应拦截器 - 解包统一响应格式 { code, msg, data }
api.interceptors.response.use(
  (response) => {
    // 解包：提取 data 字段，使 res.data 直接为实际数据
    if (response.data && 'code' in response.data) {
      response.data = response.data.data;
    }
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user: {
    id: string;
    email: string;
    role: string;
  };
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

export const authApi = {
  login: (data: LoginRequest) => api.post<AuthResponse>('/auth/login', data),
  register: (data: RegisterRequest) => api.post<AuthResponse>('/auth/register', data),
  logout: () => api.post('/auth/logout'),
  me: () => api.get('/me'),
  changePassword: (data: ChangePasswordRequest) => api.put('/auth/password', data),
};

// 物模型相关类型
export interface Property {
  id: string;
  name: string;
  dataType: 'int' | 'float' | 'bool' | 'string' | 'enum';
  unit?: string;
  min?: number;
  max?: number;
  step?: number;
  enumValues?: { value: number; label: string }[];
  accessMode: 'r' | 'rw';
}

export interface Param {
  id: string;
  name?: string;
  dataType: string;
}

export interface Event {
  id: string;
  name: string;
  params?: Param[];
}

export interface Service {
  id: string;
  name: string;
  inputParams?: Param[];
  outputParams?: Param[];
}

export interface ThingModelModule {
  id: string;
  config: {
    exposed_properties?: string[];
    exposed_services?: string[];
  };
}

export interface ThingModel {
  id: string;
  user_id: string;
  name: string;
  description: string;
  properties: Property[];
  events: Event[];
  services: Service[];
  modules: ThingModelModule[];
  created_at: string;
  updated_at: string;
}

export interface ThingModelRequest {
  name: string;
  description?: string;
  properties?: Property[];
  events?: Event[];
  services?: Service[];
  modules?: ThingModelModule[];
}

export interface ListResponse<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

export const thingModelApi = {
  list: (page = 1, pageSize = 10) =>
    api.get<ListResponse<ThingModel>>('/thing-models', { params: { page, page_size: pageSize } }),
  get: (id: string) => api.get<ThingModel>(`/thing-models/${id}`),
  create: (data: ThingModelRequest) => api.post<ThingModel>('/thing-models', data),
  update: (id: string, data: ThingModelRequest) => api.put<ThingModel>(`/thing-models/${id}`, data),
  delete: (id: string) => api.delete(`/thing-models/${id}`),
};

// 设备相关类型
export interface Device {
  id: string;
  user_id: string;
  model_id: string | null;
  model_name: string;
  name: string;
  device_secret: string;
  status: 'online' | 'offline';
  last_online_at: string | null;
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface DeviceRequest {
  name: string;
  model_id?: string;
  metadata?: Record<string, unknown>;
}

export interface DebugLog {
  id: number;
  user_id: string;
  device_id: string;
  device_name: string;
  connection_type: 'real' | 'simulated';
  action_type: 'property_set' | 'service_invoke';
  request: Record<string, unknown>;
  response?: Record<string, unknown>;
  success: boolean;
  error_message?: string;
  created_at: string;
}

export interface DeviceLatestData {
  time: string;
  payload: Record<string, unknown>;
  valid: boolean;
  errors?: Record<string, string>;
}

export interface DeviceHistoryData {
  time: string;
  payload: Record<string, unknown>;
  valid: boolean;
}

export interface AggregatedDataPoint {
  time: string;
  payload: Record<string, number>;
  max_payload: Record<string, number>;
  min_payload: Record<string, number>;
}

export interface HistoryApiResponse {
  aggregated: boolean;
  interval: string;
  data: DeviceHistoryData[] | AggregatedDataPoint[] | null;
}

export const deviceApi = {
  list: (page = 1, pageSize = 10) =>
    api.get<ListResponse<Device>>('/devices', { params: { page, page_size: pageSize } }),
  get: (id: string) => api.get<Device>(`/devices/${id}`),
  create: (data: DeviceRequest) => api.post<Device>('/devices', data),
  update: (id: string, data: DeviceRequest) => api.put<Device>(`/devices/${id}`, data),
  delete: (id: string) => api.delete(`/devices/${id}`),
  latestData: (id: string) => api.get<DeviceLatestData | null>(`/devices/${id}/data/latest`),
  dataHistory: (id: string, start: string, end: string, limit = 200) =>
    api.get<HistoryApiResponse>(`/devices/${id}/data/history`, { params: { start, end, limit } }),
  exportHistory: (id: string, start: string, end: string, limit = 1000) =>
    api.get(`/devices/${id}/data/export`, { params: { start, end, limit }, responseType: 'blob' }),
  debug: (id: string, data: { action_type: string; property_id?: string; properties?: Record<string, unknown>; service_id?: string; value?: unknown }) =>
    api.post(`/devices/${id}/debug`, data),
  connectionType: (id: string) => api.get<{ device_id: string; connection_type: 'real' | 'simulated' | 'offline' }>(`/devices/${id}/connection-type`),
  simulateOnline: (id: string) => api.post(`/devices/${id}/simulate/online`),
  simulateOffline: (id: string) => api.post(`/devices/${id}/simulate/offline`),
  simulateHeartbeat: (id: string) => api.post(`/devices/${id}/simulate/heartbeat`),
  debugLogs: (id: string, page = 1, pageSize = 20) =>
    api.get<{ list: DebugLog[]; total: number; page: number; page_size: number }>(`/devices/${id}/debug-logs`, { params: { page, page_size: pageSize } }),
};

// 统计相关类型
export interface StatsOverview {
  total_devices: number;
  online_devices: number;
  total_thing_models: number;
  today_alerts: number;
}

export const statsApi = {
  overview: () => api.get<StatsOverview>('/stats/overview'),
};

export interface Module {
  id: string;
  name: string;
  description: string;
  config_schema: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export const moduleApi = {
  list: () => api.get<Module[]>('/modules'),
  get: (id: string) => api.get<Module>(`/modules/${id}`),
};

// 审计日志相关类型
export interface AuditLog {
  id: number;
  user_id?: string;
  category: string;
  action: string;
  resource: string;
  detail?: Record<string, unknown>;
  ip: string;
  status_code: number;
  latency_ms: number;
  user_agent?: string;
  created_at: string;
}

export interface AuditLogQuery {
  category?: string;
  action?: string;
  start_time?: string;
  end_time?: string;
  page?: number;
  page_size?: number;
}

export const auditLogApi = {
  list: (params: AuditLogQuery) =>
    api.get<ListResponse<AuditLog>>('/audit-logs', { params }),
};

// 告警规则相关类型
export interface AlertRule {
  id: string;
  user_id: string;
  name: string;
  device_id: string;
  device_name: string;
  model_id: string;
  property_id: string;
  operator: '>' | '>=' | '<' | '<=' | '==' | '!=';
  threshold: number;
  severity: 'info' | 'warning' | 'critical';
  enabled: boolean;
  cooldown_minutes: number;
  created_at: string;
  updated_at: string;
}

export interface AlertRuleRequest {
  name: string;
  device_id: string;
  property_id: string;
  operator: string;
  threshold: number;
  severity: string;
  enabled: boolean;
  cooldown_minutes?: number;
}

export interface AlertLog {
  id: number;
  rule_id: string;
  user_id: string;
  device_id: string;
  device_name: string;
  property_id: string;
  property_name: string;
  operator: string;
  threshold: number;
  actual_value: number;
  severity: string;
  rule_name: string;
  acknowledged: boolean;
  acknowledged_at?: string;
  created_at: string;
}

export const alertRuleApi = {
  list: (page = 1, pageSize = 10, deviceId?: string) =>
    api.get<ListResponse<AlertRule>>('/alert-rules', { params: { page, page_size: pageSize, ...(deviceId ? { device_id: deviceId } : {}) } }),
  get: (id: string) => api.get<AlertRule>(`/alert-rules/${id}`),
  create: (data: AlertRuleRequest) => api.post<AlertRule>('/alert-rules', data),
  update: (id: string, data: AlertRuleRequest) => api.put<AlertRule>(`/alert-rules/${id}`, data),
  delete: (id: string) => api.delete(`/alert-rules/${id}`),
};

export const alertLogApi = {
  list: (page = 1, pageSize = 20, deviceId?: string) =>
    api.get<ListResponse<AlertLog>>('/alert-logs', { params: { page, page_size: pageSize, ...(deviceId ? { device_id: deviceId } : {}) } }),
  acknowledge: (id: number) => api.put(`/alert-logs/${id}/acknowledge`),
  unreadCount: () => api.get<{ count: number }>('/alert-logs/unread-count'),
};

// 定时任务相关类型
export interface ScheduledTask {
  id: string;
  user_id: string;
  device_id: string;
  device_name: string;
  name: string;
  cron_expr: string;
  action_type: 'property_set' | 'service_invoke';
  property_id?: string;
  service_id?: string;
  value?: unknown;
  enabled: boolean;
  last_run_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface ScheduledTaskRequest {
  device_id: string;
  name: string;
  cron_expr: string;
  action_type: string;
  property_id?: string;
  service_id?: string;
  value?: unknown;
  enabled: boolean;
}

export const scheduledTaskApi = {
  list: (page = 1, pageSize = 10, deviceId?: string) =>
    api.get<ListResponse<ScheduledTask>>('/scheduled-tasks', { params: { page, page_size: pageSize, ...(deviceId ? { device_id: deviceId } : {}) } }),
  get: (id: string) => api.get<ScheduledTask>(`/scheduled-tasks/${id}`),
  create: (data: ScheduledTaskRequest) => api.post<ScheduledTask>('/scheduled-tasks', data),
  update: (id: string, data: ScheduledTaskRequest) => api.put<ScheduledTask>(`/scheduled-tasks/${id}`, data),
  delete: (id: string) => api.delete(`/scheduled-tasks/${id}`),
};

// 固件相关类型
export interface Firmware {
  id: string;
  user_id: string;
  name: string;
  version: string;
  file_size: number;
  checksum: string;
  description: string;
  created_at: string;
}

export const firmwareApi = {
  list: (page = 1, pageSize = 10) =>
    api.get<ListResponse<Firmware>>('/firmwares', { params: { page, page_size: pageSize } }),
  upload: (formData: FormData) =>
    api.post<Firmware>('/firmwares', formData, { headers: { 'Content-Type': 'multipart/form-data' }, timeout: 60000 }),
  delete: (id: string) => api.delete(`/firmwares/${id}`),
};

// OTA 任务相关类型
export interface OTATask {
  id: string;
  user_id: string;
  device_id: string;
  device_name: string;
  firmware_id: string;
  firmware_version: string;
  status: 'pending' | 'pushing' | 'downloading' | 'installing' | 'completed' | 'failed' | 'cancelled';
  progress: number;
  error_msg: string;
  created_at: string;
  updated_at: string;
}

export const otaTaskApi = {
  list: (page = 1, pageSize = 10, deviceId?: string) =>
    api.get<ListResponse<OTATask>>('/ota-tasks', { params: { page, page_size: pageSize, ...(deviceId ? { device_id: deviceId } : {}) } }),
  get: (id: string) => api.get<OTATask>(`/ota-tasks/${id}`),
  create: (data: { device_id: string; firmware_id: string }) => api.post<OTATask>('/ota-tasks', data),
  cancel: (id: string) => api.put(`/ota-tasks/${id}/cancel`),
};

// 定时任务执行日志相关类型
export interface ScheduledTaskLog {
  id: number;
  task_id: string;
  user_id: string;
  device_id: string;
  device_name: string;
  task_name: string;
  action_type: 'property_set' | 'service_invoke';
  topic: string;
  payload: unknown;
  status: 'success' | 'failed';
  error: string;
  created_at: string;
}

export const scheduledTaskLogApi = {
  list: (page = 1, pageSize = 20, taskId?: string, deviceId?: string) =>
    api.get<ListResponse<ScheduledTaskLog>>('/scheduled-task-logs', {
      params: { page, page_size: pageSize, ...(taskId ? { task_id: taskId } : {}), ...(deviceId ? { device_id: deviceId } : {}) },
    }),
};

// 服务调用日志相关类型
export interface ServiceCallLog {
  id: number;
  device_id: string;
  user_id: string;
  device_name: string;
  service_id: string;
  service_name: string;
  request_id: string;
  input_params: unknown;
  output_params?: unknown;
  status: 'pending' | 'success' | 'failed' | 'timeout';
  error?: string;
  response_code?: number;
  created_at: string;
  replied_at?: string;
}

export const serviceCallLogApi = {
  list: (page = 1, pageSize = 20, deviceId?: string) =>
    api.get<ListResponse<ServiceCallLog>>('/service-call-logs', {
      params: { page, page_size: pageSize, ...(deviceId ? { device_id: deviceId } : {}) },
    }),
};

// 平台设置相关类型
export interface PlatformSettings {
  voice_mode: 'local' | 'dify';
  dify_api_url: string;
  dify_api_key: string;
  tts_provider: 'edge' | 'doubao';
  tts_doubao_app_id: string;
  tts_doubao_access_key: string;
  tts_doubao_resource_id: string;
  tts_doubao_speaker_id: string;
}

export interface UpdateSettingsRequest {
  voice_mode?: string;
  dify_api_url?: string;
  dify_api_key?: string;
  tts_provider?: string;
  tts_doubao_app_id?: string;
  tts_doubao_access_key?: string;
  tts_doubao_resource_id?: string;
  tts_doubao_speaker_id?: string;
}

export const settingsApi = {
  get: () => api.get<PlatformSettings>('/settings'),
  update: (data: UpdateSettingsRequest) => api.put<PlatformSettings>('/settings', data),
};

export const ttsApi = {
  test: (text: string) => api.post<{ audio_url: string }>('/tts/test', { text }),
};

export default api;
