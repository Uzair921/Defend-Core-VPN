import { api } from "./api";

export interface Device {
  id: string;
  user_id: string;
  name: string;
  public_key: string;
  platform: string;
  last_seen: string | null;
  status: string;
  created_at: string;
}

export interface CreateDeviceResponse {
  device: Device;
  private_key: string;
  public_key: string;
  config: string;
}

export const devicesApi = {
  list: () =>
    api.get<{ devices: Device[] }>("/devices").then((r) => r.data.devices),

  create: (name: string, platform: string) =>
    api
      .post<CreateDeviceResponse>("/devices", { name, platform })
      .then((r) => r.data),

  get: (id: string) =>
    api.get<Device>(`/devices/${id}`).then((r) => r.data),

  delete: (id: string) =>
    api.delete(`/devices/${id}`).then((r) => r.data),
};
