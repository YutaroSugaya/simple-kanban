import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';
import { useAuth } from '../contexts/AuthContext';
import { boardApi, calendarApi } from '../lib/api';
import Calendar from '../components/Calendar';
import Timer from '../components/Timer';
import * as Types from '../types';

/**
 * カレンダーページコンポーネント
 */
const CalendarPage: React.FC = () => {
  const { user, logout } = useAuth();
  const [selectedTask, setSelectedTask] = useState<Types.Task | null>(null);
  const [boards, setBoards] = useState<Types.Board[]>([]);
  const [loading, setLoading] = useState(true);
  const [calendarKey, setCalendarKey] = useState(0); // カレンダーコンポーネントの強制再レンダリング用
  const [timerTaskId, setTimerTaskId] = useState<number | undefined>(undefined); // タイマー開始用のタスクID

  // ボードとタスクの取得
  useEffect(() => {
    const fetchBoards = async () => {
      try {
        setLoading(true);
        const boardsData = await boardApi.getBoardsWithColumns();
        setBoards(boardsData);
      } catch (error) {
        console.error('Failed to fetch boards:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchBoards();
  }, []);

  const handleTaskSelect = (task: Types.Task) => {
    setSelectedTask(task);
  };

  const handleTimerStart = (taskId: number) => {
    setTimerTaskId(taskId);
    // 少し待ってからリセット（一度だけトリガーするため）
    setTimeout(() => setTimerTaskId(undefined), 100);
  };

  // タスクをクリックした時にカレンダーに追加
  const handleTaskClick = async (task: Types.Task) => {
    // タスクを選択
    handleTaskSelect(task);

    try {
      // 現在の時間でカレンダーに追加
      const now = new Date();
      const startDateTime = new Date(now);
      // 分を10分刻みに調整（例：14:23 → 14:20）
      startDateTime.setMinutes(Math.floor(startDateTime.getMinutes() / 10) * 10, 0, 0);
      
      const endDateTime = new Date(startDateTime);
      const durationMinutes = task.estimated_time || 30;
      endDateTime.setMinutes(endDateTime.getMinutes() + durationMinutes);

      console.log('CalendarPage: イベント作成リクエスト', {
        taskId: task.id,
        title: task.title,
        start: startDateTime.toISOString(),
        end: endDateTime.toISOString()
      });

      await calendarApi.createEventFromTask(task.id, startDateTime.toISOString(), endDateTime.toISOString());
      
      console.log('CalendarPage: イベント作成成功、カレンダー更新開始');
      
      // 成功メッセージを表示
      alert(`${task.title} をカレンダーに追加しました！`);
      
      // カレンダーを更新
      setCalendarKey(prev => prev + 1);
      
      // 少し待ってからカレンダーを更新（イベントがDBに保存されるのを待つ）
      setTimeout(() => {
        console.log('CalendarPage: 遅延カレンダー更新実行');
        setCalendarKey(prev => prev + 1);
      }, 1000);
    } catch (error: unknown) {
      console.error('Failed to add task to calendar:', error);
      
      let errorMessage = 'タスクの追加に失敗しました';
      if (error && typeof error === 'object' && 'response' in error) {
        const axiosError = error as { response?: { data?: { error?: string } } };
        if (axiosError.response?.data?.error) {
          errorMessage = axiosError.response.data.error;
        }
      } else if (error instanceof Error && error.message) {
        errorMessage = error.message;
      }
      
      alert(errorMessage);
    }
  };

  // ドラッグ&ドロップ処理
  const onDragEnd = async (result: Types.DragResult) => {
    const { destination, draggableId } = result;

    if (!destination) return;

    // カレンダーにドロップされた場合
    if (destination.droppableId.startsWith('calendar-slot-')) {
      // droppableId 形式: "calendar-slot-YYYY-MM-DD-HH:MM"
      const parts = destination.droppableId.split('-');
      const time = parts[parts.length - 1];
      const date = parts.slice(2, parts.length - 1).join('-');
      const taskId = parseInt(draggableId.replace('task-', ''));
      
      try {
        // まずタスクの詳細を取得して目標時間を確認
        const task = boards
          .flatMap(board => board.columns || [])
          .flatMap(column => column.tasks || [])
          .find(task => task.id === taskId);

        // ドロップされた日付と時間を使用してイベントを作成
        const startDateTime = new Date(`${date}T${time}:00`);
        const endDateTime = new Date(startDateTime);
        
        // タスクの目標時間があれば使用、なければデフォルト30分
        const durationMinutes = task?.estimated_time || 30;
        endDateTime.setMinutes(endDateTime.getMinutes() + durationMinutes);

        await calendarApi.createEventFromTask(taskId, startDateTime.toISOString(), endDateTime.toISOString());
        
        // 成功メッセージは非表示
        
        // カレンダーコンポーネントを強制再レンダリング
        setCalendarKey(prev => prev + 1);
        
        // 少し待ってからカレンダーを更新（イベントがDBに保存されるのを待つ）
        setTimeout(() => {
          setCalendarKey(prev => prev + 1);
        }, 1000);
      } catch (error: unknown) {
        
        let errorMessage = 'タスクの追加に失敗しました';
        if (error && typeof error === 'object' && 'response' in error) {
          const axiosError = error as { response?: { data?: { error?: string }; status?: number }; message?: string };
          
          
          if (axiosError.response?.data?.error) {
            errorMessage = axiosError.response.data.error;
          } else if (axiosError.message) {
            errorMessage = axiosError.message;
          }
        } else if (error instanceof Error && error.message) {
          errorMessage = error.message;
        }
        
        alert(errorMessage);
      }
    }
  };



  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* ヘッダー */}
      <nav className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-4">
              <Link
                to="/"
                className="text-gray-500 hover:text-gray-700"
              >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                </svg>
              </Link>
              <h1 className="text-xl font-semibold text-gray-900">カレンダー</h1>
            </div>
            <div className="flex items-center space-x-4">
              <Link
                to="/settings"
                className="text-sm text-gray-600 hover:text-gray-800"
              >
                設定
              </Link>
              <span className="text-sm text-gray-600">
                {user?.email}
              </span>
              <button
                onClick={logout}
                className="text-sm text-gray-600 hover:text-gray-800"
              >
                ログアウト
              </button>
            </div>
          </div>
        </div>
      </nav>

      {/* メインコンテンツ */}
      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          <DragDropContext onDragEnd={onDragEnd}>
            <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
              {/* カレンダー */}
              <div className="lg:col-span-3">
                <Calendar
                  key={calendarKey}
                  refreshKey={calendarKey}
                  onTaskSelect={handleTaskSelect}
                  onTimerStart={handleTimerStart}
                />
              </div>

              {/* サイドバー */}
              <div className="lg:col-span-1 space-y-6">
                {/* タスク一覧（ドラッグ可能） */}
                <div className="bg-white rounded-lg shadow-sm border p-4">
                  <h3 className="text-lg font-medium text-gray-900 mb-4">タスク一覧</h3>
                  <p className="text-sm text-gray-600 mb-4">
                    ドラッグ&ドロップで指定の時間に追加
                  </p>
                  
                  {boards.length === 0 ? (
                    <div className="text-center py-4">
                      <p className="text-gray-500">ボードがありません</p>
                      <Link to="/" className="text-blue-600 hover:text-blue-800 text-sm">
                        ボードを作成
                      </Link>
                    </div>
                  ) : (
                    <div className="space-y-2 max-h-64 overflow-y-auto">
                      {boards.map(board => (
                        <div key={board.id} className="border rounded p-2">
                          <h4 className="font-medium text-sm text-gray-700 mb-2">{board.name}</h4>
                          {board.columns?.map(column => (
                            <div key={column.id} className="ml-2 mb-2">
                              <h5 className="text-xs text-gray-500 mb-1">{column.title}</h5>
                              <Droppable droppableId={`task-list-${column.id}`} isDropDisabled={true}>
                                {(provided) => (
                                  <div
                                    ref={provided.innerRef}
                                    {...provided.droppableProps}
                                    className="space-y-1"
                                  >
                                    {column.tasks?.map((task, index) => (
                                      <Draggable
                                        key={task.id}
                                        draggableId={`task-${task.id}`}
                                        index={index}
                                      >
                                        {(provided, snapshot) => (
                                          <div
                                            ref={provided.innerRef}
                                            {...provided.draggableProps}
                                            {...provided.dragHandleProps}
                                            className={`p-2 text-xs rounded cursor-move relative ${
                                              snapshot.isDragging
                                                ? 'bg-blue-100 border-2 border-blue-300'
                                                : selectedTask?.id === task.id 
                                                ? 'bg-blue-50 border border-blue-300'
                                                : 'bg-gray-50 hover:bg-gray-100'
                                            }`}
                                            onClick={() => handleTaskClick(task)}
                                          >
                                            <div className="font-medium truncate">{task.title}</div>
                                            {task.estimated_time && (
                                              <div className="text-gray-500">
                                                予定: {task.estimated_time}分
                                              </div>
                                            )}
                                            {/* "+" ボタンは不要のため削除 */}
                                          </div>
                                        )}
                                      </Draggable>
                                    ))}
                                    {provided.placeholder}
                                  </div>
                                )}
                              </Droppable>
                            </div>
                          ))}
                        </div>
                      ))}
                    </div>
                  )}
                </div>

                {/* タイマー */}
                <div className="bg-white rounded-lg shadow-sm border p-4">
                  <h3 className="text-lg font-medium text-gray-900 mb-4">タイマー</h3>
                  <Timer 
                    task={selectedTask || undefined}
                    externalTaskId={timerTaskId}
                    onTimerComplete={() => {
                      alert('タイマーが完了しました！');
                    }}
                    onTimerStop={(actualDuration) => {
                      console.log('タイマー停止:', actualDuration);
                    }}
                  />
                </div>

                {/* タスク詳細 */}
                {selectedTask && (
                  <div className="bg-white rounded-lg shadow-sm border p-4">
                    <h3 className="text-lg font-medium text-gray-900 mb-4">選択されたタスク</h3>
                    <div className="space-y-3">
                      <div>
                        <label className="block text-sm font-medium text-gray-700">タイトル</label>
                        <p className="mt-1 text-sm text-gray-900">{selectedTask.title}</p>
                      </div>
                      {selectedTask.description && (
                        <div>
                          <label className="block text-sm font-medium text-gray-700">説明</label>
                          <p className="mt-1 text-sm text-gray-900">{selectedTask.description}</p>
                        </div>
                      )}
                      {selectedTask.estimated_time && (
                        <div>
                          <label className="block text-sm font-medium text-gray-700">目標時間</label>
                          <p className="mt-1 text-sm text-gray-900">{selectedTask.estimated_time}分</p>
                        </div>
                      )}
                      {selectedTask.actual_time && (
                        <div>
                          <label className="block text-sm font-medium text-gray-700">実際の時間</label>
                          <p className="mt-1 text-sm text-gray-900">{selectedTask.actual_time}分</p>
                        </div>
                      )}
                      <div>
                        <label className="block text-sm font-medium text-gray-700">ステータス</label>
                        <p className={`mt-1 text-sm ${
                          selectedTask.is_completed ? 'text-green-600' : 'text-yellow-600'
                        }`}>
                          {selectedTask.is_completed ? '完了' : '進行中'}
                        </p>
                      </div>
                    </div>
                  </div>
                )}

                {/* 使用方法 */}
                <div className="bg-white rounded-lg shadow-sm border p-4">
                  <h3 className="text-lg font-medium text-gray-900 mb-4">使い方</h3>
                  <div className="space-y-2 text-sm text-gray-600">
                    <p>• タスクをクリックで現在の時間にカレンダー追加</p>
                    <p>• ドラッグ&ドロップで指定の時間に追加</p>
                    <p>• カレンダー内のタスクをクリックしてタイマー開始</p>
                    <p>• 「日」「週」ボタンで表示を切り替え</p>
                    <p>• 設定ページで表示時間をカスタマイズ</p>
                  </div>
                </div>

                {/* ショートカット */}
                <div className="bg-white rounded-lg shadow-sm border p-4">
                  <h3 className="text-lg font-medium text-gray-900 mb-4">クイックアクション</h3>
                  <div className="space-y-2">
                    <Link
                      to="/"
                      className="w-full btn-secondary text-center block"
                    >
                      カンバンボードへ
                    </Link>
                    <Link
                      to="/settings"
                      className="w-full btn-secondary text-center block"
                    >
                      カレンダー設定
                    </Link>
                  </div>
                </div>
              </div>
            </div>
          </DragDropContext>
        </div>
      </div>
    </div>
  );
};

export default CalendarPage; 