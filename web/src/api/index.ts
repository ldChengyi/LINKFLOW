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

export const authApi = {
  login: (data: LoginRequest) => api.post<AuthResponse>('/auth/login', data),
  register: (data: RegisterRequest) => api.post<AuthResponse>('/auth/register', data),
  logout: () => api.post('/auth/logout'),
  me: () => api.get('/me'),
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

export const deviceApi = {
  list: (page = 1, pageSize = 10) =>
    api.get<ListResponse<Device>>('/devices', { params: { page, page_size: pageSize } }),
  get: (id: string) => api.get<Device>(`/devices/${id}`),
  create: (data: DeviceRequest) => api.post<Device>('/devices', data),
  update: (id: string, data: DeviceRequest) => api.put<Device>(`/devices/${id}`, data),
  delete: (id: string) => api.delete(`/devices/${id}`),
  latestData: (id: string) => api.get<DeviceLatestData | null>(`/devices/${id}/data/latest`),
  dataHistory: (id: string, start: string, end: string, limit = 200) =>
    api.get<DeviceHistoryData[]>(`/devices/${id}/data/history`, { params: { start, end, limit } }),
  debug: (id: string, data: { action_type: string; property_id?: string; properties?: Record<string, unknown>; service_id?: string; value?: unknown }) =>
    api.post(`/devices/${id}/debug`, data),
  connectionType: (id: string) => api.get<{ device_id: string; connection_type: 'real' | 'simulated' | 'offline' }>(`/devices/${id}/connection-type`),
  simulateOnline: (id: string) => api.post(`/devices/${id}/simulate/online`),
  simulateOffline: (id: string) => api.post(`/devices/${id}/simulate/offline`),
  simulateHeartbeat: (id: string) => api.post(`/devices/${id}/simulate/heartbeat`),
};

// 统计相关类型
export interface StatsOverview {
  total_devices: number;
  online_devices: number;
  total_thing_models: number;
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

export default api;
