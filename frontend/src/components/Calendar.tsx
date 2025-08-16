import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { Droppable } from '@hello-pangea/dnd';
import { calendarApi, timerApi, taskApi } from '../lib/api';
import { useAuth } from '../contexts/AuthContext';
import * as Types from '../types';

interface CalendarProps {
  onTaskSelect?: (task: Types.Task) => void;
  refreshKey?: number; // イベント再取得用のキー
  onTimerStart?: (taskId: number) => void; // タイマー開始のコールバック
}

/**
 * カレンダーコンポーネント
 * 日単位・週間単位の切り替え、10分刻み表示、ドラッグ&ドロップ対応
 */
const Calendar: React.FC<CalendarProps> = ({ onTaskSelect, refreshKey, onTimerStart }) => {
  const { isAuthenticated, loading: authLoading } = useAuth();
  const [currentDate, setCurrentDate] = useState(() => {
    // 明示的に今日の日付を設定（タイムゾーンの問題を避ける）
    const now = new Date();
    const todayLocal = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    return todayLocal;
  });
  const [view, setView] = useState<Types.CalendarView>('day');
  const [events, setEvents] = useState<Types.CalendarEvent[]>([]);
  const [settings, setSettings] = useState<Types.CalendarSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTimer, setActiveTimer] = useState<Types.TimerSession | null>(null);

  // 表示期間を取得
  const getDateRange = useCallback(() => {
    // 現在選択されている日付をベースに範囲を計算
    // 日付オブジェクトを直接複製して、時間情報を正確に処理
    const baseDate = new Date(currentDate.getTime());
    

    if (view === 'day') {
      // その日の00:00:00から23:59:59まで（ローカルタイム）
      const start = new Date(baseDate);
      start.setHours(0, 0, 0, 0);
      
      const end = new Date(baseDate);
      end.setHours(23, 59, 59, 999);

      

      return { start, end };
    } else { // week
      // 週の最初の日（日曜日）を基準にする
      const start = new Date(baseDate);
      const dayOfWeek = start.getDay();
      start.setDate(start.getDate() - dayOfWeek);
      start.setHours(0, 0, 0, 0);
      
      const end = new Date(start);
      end.setDate(start.getDate() + 6);
      end.setHours(23, 59, 59, 999);

      

      return { start, end };
    }
  }, [currentDate, view]);

  // イベントデータの取得
  const fetchEvents = useCallback(async () => {
    const { start, end } = getDateRange();
    try {
      const eventsData = await calendarApi.getEvents(start.toISOString(), end.toISOString());
      setEvents(eventsData);
    } catch {
      // no-op
    }
  }, [getDateRange]);

  // カレンダー設定とイベントの取得
  useEffect(() => {
    // 認証が完了していない場合は何もしない
    if (authLoading || !isAuthenticated) return;

    const fetchData = async () => {
      try {
        setLoading(true);
        const settingsData = await calendarApi.getSettings();
        setSettings(settingsData);
        
        // アクティブタイマーを取得（404は既にapi.tsで処理済み）
        try {
          const timerData = await timerApi.getActiveTimer();
          setActiveTimer(timerData);
        } catch {
          setActiveTimer(null);
        }
        
        // イベントデータを取得
        await fetchEvents();
      } catch {
        // no-op
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [currentDate, view, refreshKey, fetchEvents, isAuthenticated, authLoading]); // 認証状態を依存配列に追加

  // 時間スロットを生成
  const timeSlots = useMemo(() => {
    if (!settings) return [];

    const isWeekend = currentDate.getDay() === 0 || currentDate.getDay() === 6;
    const startTime = isWeekend ? settings.weekend_start_time : settings.weekday_start_time;
    const endTime = isWeekend ? settings.weekend_end_time : settings.weekday_end_time;
    const slotDuration = settings.time_slot_duration;

    const slots: Types.TimeSlot[] = [];
    const [startHour, startMinute] = startTime.split(':').map(Number);
    const [endHour, endMinute] = endTime.split(':').map(Number);

    let currentHour = startHour;
    let currentMinute = startMinute;

    while (currentHour < endHour || (currentHour === endHour && currentMinute < endMinute)) {
      const timeString = `${currentHour.toString().padStart(2, '0')}:${currentMinute.toString().padStart(2, '0')}`;
      slots.push({
        time: timeString,
        hour: currentHour,
        minute: currentMinute,
      });

      currentMinute += slotDuration;
      if (currentMinute >= 60) {
        currentHour += Math.floor(currentMinute / 60);
        currentMinute = currentMinute % 60;
      }
    }

    return slots;
  }, [settings, currentDate]);

  // 日付配列を生成（週表示用）
  const weekDays = useMemo(() => {
    if (view === 'day') return [currentDate];

    const days = [];
    const startOfWeek = new Date(currentDate);
    const dayOfWeek = startOfWeek.getDay();
    startOfWeek.setDate(startOfWeek.getDate() - dayOfWeek);

    for (let i = 0; i < 7; i++) {
      const day = new Date(startOfWeek);
      day.setDate(startOfWeek.getDate() + i);
      days.push(day);
    }

    return days;
  }, [currentDate, view]);

  // ドラッグ&ドロップ処理


  // 日付ナビゲーション
  const navigateDate = (direction: 'prev' | 'next') => {
    const newDate = new Date(currentDate);
    if (view === 'day') {
      newDate.setDate(newDate.getDate() + (direction === 'next' ? 1 : -1));
    } else {
      newDate.setDate(newDate.getDate() + (direction === 'next' ? 7 : -7));
    }
    setCurrentDate(newDate);
  };

  // タイマー開始
  const startTimer = async (taskId: number) => {
    try {
      // タスク詳細を取得
      const task = await taskApi.getTask(taskId);
      
      // 外部のタイマーコンポーネントに開始を通知
      if (onTimerStart) {
        onTimerStart(taskId);
      }
      
      // タスク詳細を取得して選択状態を更新
      if (onTaskSelect) {
        onTaskSelect(task);
      }
    } catch (error) {
      console.error('Failed to start timer:', error);
      alert('タイマーの開始に失敗しました');
    }
  };

  // タイマー停止
  const stopTimer = async () => {
    if (!activeTimer) return;
    
    try {
      await timerApi.stopTimer(activeTimer.id!);
      setActiveTimer(null);
    } catch (error) {
      console.error('Failed to stop timer:', error);
      alert('タイマーの停止に失敗しました');
    }
  };

  // タスク詳細を取得して選択
  const handleTaskClick = async (event: Types.CalendarEvent) => {
    if (event.task_id && onTaskSelect) {
      try {
        const task = await taskApi.getTask(event.task_id);
        onTaskSelect(task);
      } catch {
        // no-op
      }
    }
  };

  // カレンダーのイベント操作（完了/削除）
  const completeTaskFromEvent = async (event: Types.CalendarEvent) => {
    if (!event.task_id) return;
    try {
      const current = await taskApi.getTask(event.task_id);
      await taskApi.updateTask(event.task_id, { is_completed: !current.is_completed });
      // 再取得して反映（完了タスクも表示は継続）
      await fetchEvents();
    } catch {
      // no-op
    }
  };

  const deleteEvent = async (eventId?: number) => {
    if (!eventId) return;
    try {
      await calendarApi.deleteEvent(eventId);
      await fetchEvents();
    } catch {
      // no-op
    }
  };

  // 指定時間のイベントを取得（重複を除去して統合）
  const getEventsForSlot = (date: Date, timeSlot: Types.TimeSlot) => {
    
    const slotStart = new Date(date);
    slotStart.setHours(timeSlot.hour, timeSlot.minute, 0, 0);
    const slotEnd = new Date(slotStart);
    slotEnd.setMinutes(slotEnd.getMinutes() + (settings?.time_slot_duration || 30));

    const slotEvents = events.filter(event => {
      const eventStart = new Date(event.start);
      const eventEnd = new Date(event.end);
      const isInSlot = eventStart < slotEnd && eventEnd > slotStart;
      return isInSlot;
    });

    // タスクベースのイベントを統合（同じタスクIDのイベントを1つにまとめる）
    const taskBasedEvents = slotEvents.filter(event => event.is_task_based);
    const otherEvents = slotEvents.filter(event => !event.is_task_based);
    
    // タスクIDでグループ化して統合
    const taskGroups = new Map<number, Types.CalendarEvent>();
    taskBasedEvents.forEach(event => {
      if (event.task_id) {
        if (!taskGroups.has(event.task_id)) {
          taskGroups.set(event.task_id, event);
        } else {
          // 既存のイベントと時間範囲を統合
          const existing = taskGroups.get(event.task_id)!;
          const existingStart = new Date(existing.start);
          const existingEnd = new Date(existing.end);
          const eventStart = new Date(event.start);
          const eventEnd = new Date(event.end);
          
          const newStart = existingStart < eventStart ? existingStart : eventStart;
          const newEnd = existingEnd > eventEnd ? existingEnd : eventEnd;
          
          taskGroups.set(event.task_id, {
            ...existing,
            start: newStart.toISOString(),
            end: newEnd.toISOString()
          });
        }
      }
    });

    return [...taskGroups.values(), ...otherEvents];
  };

  // イベントの表示時間を計算
  const getEventDisplayTime = (event: Types.CalendarEvent, date: Date, timeSlot: Types.TimeSlot) => {
    const slotStart = new Date(date);
    slotStart.setHours(timeSlot.hour, timeSlot.minute, 0, 0);
    const slotEnd = new Date(slotStart);
    slotEnd.setMinutes(slotEnd.getMinutes() + (settings?.time_slot_duration || 30));
    
    const eventStart = new Date(event.start);
    const eventEnd = new Date(event.end);
    
    // イベントがこのスロットで開始するかチェック
    const startsInSlot = eventStart >= slotStart && eventStart < slotEnd;
    
    return {
      startsInSlot,
      isActive: eventStart < slotEnd && eventEnd > slotStart
    };
  };

  if (authLoading || loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  // 認証されていない場合は何も表示しない（ProtectedRouteで処理される）
  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="bg-white rounded-lg shadow-sm border">
      {/* ヘッダー */}
      <div className="p-4 border-b border-gray-200">
        <div className="flex justify-between items-center">
          <div className="flex items-center space-x-4">
            {/* 日付ナビゲーション */}
            <div className="flex items-center space-x-2">
              <button
                onClick={() => navigateDate('prev')}
                className="p-1 hover:bg-gray-100 rounded"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                </svg>
              </button>
              <h2 className="text-lg font-semibold">
                {view === 'day' 
                  ? currentDate.toLocaleDateString('ja-JP', { year: 'numeric', month: 'long', day: 'numeric', weekday: 'long' })
                  : `${weekDays[0].toLocaleDateString('ja-JP', { month: 'short', day: 'numeric' })} - ${weekDays[6].toLocaleDateString('ja-JP', { month: 'short', day: 'numeric' })}`
                }
              </h2>
              <button
                onClick={() => navigateDate('next')}
                className="p-1 hover:bg-gray-100 rounded"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </button>
            </div>

            {/* 今日ボタン */}
            <button
              onClick={() => {
                const today = new Date();
                // タイムゾーンの問題を避けるため、明示的に今日の日付を設定
                const todayLocal = new Date(today.getFullYear(), today.getMonth(), today.getDate());
                console.log('Calendar: 今日ボタンクリック', {
                  now: today.toISOString(),
                  todayLocal: todayLocal.toISOString(),
                  localString: today.toLocaleDateString('ja-JP')
                });
                setCurrentDate(todayLocal);
              }}
              className="px-3 py-1 text-sm bg-blue-100 text-blue-700 rounded hover:bg-blue-200"
            >
              今日
            </button>
          </div>

          <div className="flex items-center space-x-4">
            {/* アクティブタイマー表示 */}
            {activeTimer && (
              <div className="flex items-center space-x-2 px-3 py-1 bg-green-100 text-green-700 rounded">
                <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                <span className="text-sm font-medium">{activeTimer.task?.title || 'タイマー実行中'}</span>
                <button
                  onClick={stopTimer}
                  className="text-xs px-2 py-1 bg-green-500 text-white rounded hover:bg-green-600"
                >
                  停止
                </button>
              </div>
            )}

            {/* ビュー切り替え */}
            <div className="flex bg-gray-100 rounded-lg p-1">
              <button
                onClick={() => setView('day')}
                className={`px-3 py-1 text-sm rounded ${
                  view === 'day' ? 'bg-white shadow-sm' : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                日
              </button>
              <button
                onClick={() => setView('week')}
                className={`px-3 py-1 text-sm rounded ${
                  view === 'week' ? 'bg-white shadow-sm' : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                週
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* カレンダーグリッド */}
      <div className="overflow-auto max-h-96">
        <div className="grid grid-cols-1" style={{ gridTemplateColumns: view === 'week' ? 'auto repeat(7, 1fr)' : 'auto 1fr' }}>
          {/* ヘッダー行（週表示時の日付） */}
          {view === 'week' && (
            <>
              <div className="p-2 border-b border-gray-200"></div>
              {weekDays.map(day => (
                <div key={day.toDateString()} className="p-2 border-b border-l border-gray-200 text-center">
                  <div className="text-sm font-medium">{day.toLocaleDateString('ja-JP', { weekday: 'short' })}</div>
                  <div className="text-lg">{day.getDate()}</div>
                </div>
              ))}
            </>
          )}

          {/* 時間スロット */}
          {timeSlots.map((timeSlot) => (
            <React.Fragment key={timeSlot.time}>
              {/* 時間ラベル */}
              <div className="p-2 border-b border-gray-200 text-right text-sm text-gray-600 w-16">
                {timeSlot.time}
              </div>

              {/* 各日のタイムスロット */}
              {weekDays.map(day => {
                const slotEvents = getEventsForSlot(day, timeSlot);
                // ローカル日付キー (YYYY-MM-DD) を生成してUTC変換による日付ズレを防止
                const dateKey = `${day.getFullYear()}-${(day.getMonth() + 1)
                  .toString()
                  .padStart(2, '0')}-${day.getDate().toString().padStart(2, '0')}`;
                const slotId = `calendar-slot-${dateKey}-${timeSlot.time}`;
                
                return (
                  <Droppable key={slotId} droppableId={slotId}>
                    {(provided, snapshot) => (
                      <div
                        ref={provided.innerRef}
                        {...provided.droppableProps}
                        className={`h-12 p-1 border-b border-l border-gray-200 relative ${
                          snapshot.isDraggingOver ? 'bg-blue-50' : 'hover:bg-gray-50'
                        }`}
                      >
                        {/* イベント表示 */}
                        {slotEvents.map((event) => {
                          const displayTime = getEventDisplayTime(event, day, timeSlot);
                          
                          // タスクベースのイベントは開始スロットでのみ表示
                          if (event.is_task_based && !displayTime.startsInSlot) {
                            return null;
                          }
                          
                          // イベントの表示の高さを計算（目標時間に基づく）
                          const eventStart = new Date(event.start);
                          const eventEnd = new Date(event.end);
                          const eventDurationMinutes = Math.round((eventEnd.getTime() - eventStart.getTime()) / (1000 * 60));
                          const slotDuration = settings?.time_slot_duration || 30;
                          const heightSlots = Math.max(1, Math.ceil(eventDurationMinutes / slotDuration));
                          const heightPx = heightSlots * 48 - 4; // 48px per slot, 4px for margins
                          
                          const isCompleted = !!(event as Types.CalendarEvent).task?.is_completed;
                          return (
                            <div
                              key={event.id}
                              className={`text-xs p-1 rounded mb-1 cursor-pointer relative ${
                                event.is_task_based ? 'bg-green-100 text-green-800 hover:bg-green-200' : 'bg-blue-100 text-blue-800 hover:bg-blue-200'
                              } ${event.is_task_based && isCompleted ? 'opacity-60 line-through' : ''}`}
                              style={{ 
                                height: event.is_task_based ? `${heightPx}px` : 'auto',
                                zIndex: 10
                              }}
                              onClick={() => handleTaskClick(event)}
                            >
                              <div className="font-medium truncate flex items-center justify-between">
                                <span>{event.title}</span>
                                {event.is_task_based && (
                                  <div className="flex items-center gap-1 ml-2">
                                    <button
                                      onClick={(e) => {
                                        e.stopPropagation();
                                        completeTaskFromEvent(event);
                                      }}
                                      className={`hover:text-green-900 ${ isCompleted ? 'text-gray-600' : 'text-green-700'}`}
                                      title={isCompleted ? '未完了にする' : '完了にする'}
                                      aria-label="toggle-complete-event"
                                      type="button"
                                    >
                                      {isCompleted ? '↺' : '✓'}
                                    </button>
                                    <button
                                      onClick={(e) => {
                                        e.stopPropagation();
                                        deleteEvent(Number(event.id));
                                      }}
                                      className="text-red-700 hover:text-red-900"
                                      title="イベント削除"
                                      aria-label="delete-event"
                                      type="button"
                                    >
                                      ×
                                    </button>
                                  </div>
                                )}
                              </div>
                              {event.is_task_based && (
                                <button
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    if (event.task_id) {
                                      startTimer(event.task_id);
                                    }
                                  }}
                                  className="mt-1 text-xs px-1 py-0.5 bg-green-500 text-white rounded hover:bg-green-600 disabled:opacity-50 transition-colors"
                                  disabled={!!activeTimer}
                                >
                                  {activeTimer && activeTimer.task_id === event.task_id ? (
                                    <span className="flex items-center">
                                      <div className="w-2 h-2 bg-white rounded-full animate-pulse mr-1"></div>
                                      実行中
                                    </span>
                                  ) : (
                                    'タイマー開始'
                                  )}
                                </button>
                              )}
                            </div>
                          );
                        })}
                        {provided.placeholder}
                      </div>
                    )}
                  </Droppable>
                );
              })}
            </React.Fragment>
          ))}
        </div>
      </div>
    </div>
  );
};

export default Calendar; 