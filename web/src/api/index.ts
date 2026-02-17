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

export const deviceApi = {
  list: (page = 1, pageSize = 10) =>
    api.get<ListResponse<Device>>('/devices', { params: { page, page_size: pageSize } }),
  get: (id: string) => api.get<Device>(`/devices/${id}`),
  create: (data: DeviceRequest) => api.post<Device>('/devices', data),
  update: (id: string, data: DeviceRequest) => api.put<Device>(`/devices/${id}`, data),
  delete: (id: string) => api.delete(`/devices/${id}`),
  latestData: (id: string) => api.get<DeviceLatestData | null>(`/devices/${id}/data/latest`),
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

export default api;
