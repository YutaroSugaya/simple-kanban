// ユーザー関連の型定義
export interface User {
  id: string;
  email: string;
  created_at: string;
  updated_at: string;
}

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
  user: User;
}

// タスク関連の型定義（拡張）
export interface Task {
  id: number;
  column_id: number;
  title: string;
  description?: string;
  order: number;
  assignee_id?: string;
  due_date?: string;
  created_at: string;
  updated_at: string;
  // 新機能用フィールド
  estimated_time?: number; // 目標時間（分）
  actual_time?: number; // 実際にかかった時間（分）
  is_completed?: boolean; // 完了状態
  scheduled_start?: string; // スケジュール開始時刻
  scheduled_end?: string; // スケジュール終了時刻
  calendar_date?: string; // カレンダー配置日
}

export interface CreateTaskRequest {
  column_id: number;
  title: string;
  description?: string;
  order: number;
  assignee_id?: string;
  due_date?: string;
  // 新機能用フィールド
  estimated_time?: number;
  is_completed?: boolean;
  scheduled_start?: string;
  scheduled_end?: string;
  calendar_date?: string;
}

export interface UpdateTaskRequest {
  title?: string;
  description?: string;
  order?: number;
  assignee_id?: string;
  due_date?: string;
  // 新機能用フィールド
  estimated_time?: number;
  is_completed?: boolean;
  scheduled_start?: string;
  scheduled_end?: string;
  calendar_date?: string;
}

export interface MoveTaskRequest {
  new_column_id: number;
  new_order: number;
}

// カラム関連の型定義
export interface Column {
  id: number;
  board_id: number;
  title: string;
  order: number;
  created_at: string;
  updated_at: string;
  tasks?: Task[];
}

export interface CreateColumnRequest {
  title: string;
  order: number;
}

export interface UpdateColumnRequest {
  title: string;
  order: number;
}

// ボード関連の型定義
export interface Board {
  id: number;
  name: string;
  owner_id: string;
  created_at: string;
  updated_at: string;
  columns?: Column[];
}

export interface CreateBoardRequest {
  name: string;
}

export interface UpdateBoardRequest {
  name: string;
}

// API レスポンス関連の型定義
export interface ApiResponse<T> {
  data: T;
  message?: string;
}

export interface ApiError {
  error: string;
  message?: string;
}

// ドラッグ&ドロップ関連の型定義
export interface DragResult {
  draggableId: string;
  type: string;
  source: {
    droppableId: string;
    index: number;
  };
  destination: {
    droppableId: string;
    index: number;
  } | null;
}

// フォーム関連の型定義
export interface FormError {
  field: string;
  message: string;
}

// アプリケーション状態の型定義
export interface AppState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
}

// カレンダー関連の型定義
export interface CalendarEvent {
  id: number;
  task_id?: number;
  title: string; // ISO 8601 format
  start: string; // ISO 8601 format
  end: string;
  color?: string;
  is_task_based: boolean;
  task?: Task; // サーバーから付与される関連タスク（任意）
}

export interface CalendarSettings {
  id?: number;
  user_id: string;
  weekday_start_time: string; // "09:00"
  weekday_end_time: string; // "18:00"
  weekend_start_time: string; // "10:00"
  weekend_end_time: string; // "16:00"
  time_slot_duration: number; // 10分刻み
  created_at?: string;
  updated_at?: string;
}

export interface CreateCalendarSettingsRequest {
  weekday_start_time: string;
  weekday_end_time: string;
  weekend_start_time: string;
  weekend_end_time: string;
  time_slot_duration: number;
}

export interface UpdateCalendarSettingsRequest {
  weekday_start_time: string;
  weekday_end_time: string;
  weekend_start_time: string;
  weekend_end_time: string;
  time_slot_duration: number;
}

// タイマー関連の型定義
export interface TimerSession {
  id?: number;
  task_id: number;
  user_id: string;
  start_time: string;
  end_time?: string;
  duration: number; // 秒
  is_active: boolean;
  created_at?: string;
  updated_at?: string;
  // リレーション
  task?: Task;
}

export interface CreateTimerSessionRequest {
  task_id: number;
  duration: number;
}

export interface UpdateTimerSessionRequest {
  end_time?: string;
  duration: number;
  is_active: boolean;
}

// カレンダービュー関連の型定義
export type CalendarView = 'day' | 'week';

export interface TimeSlot {
  time: string; // "09:00"
  hour: number;
  minute: number;
}

export interface CalendarDay {
  date: string; // "2024-01-15"
  dayOfWeek: number; // 0-6 (Sunday-Saturday)
  timeSlots: TimeSlot[];
  events: CalendarEvent[];
} 