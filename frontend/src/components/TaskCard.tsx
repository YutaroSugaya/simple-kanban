import React, { useState } from 'react';
import { Draggable } from '@hello-pangea/dnd';
import * as Types from '../types';
import { taskApi, timerApi } from '../lib/api';
import { useForm } from 'react-hook-form';
import { format, parseISO } from 'date-fns';

interface TaskCardProps {
  task: Types.Task;
  index: number;
  onUpdate: (updatedTask: Types.Task) => void;
  onDelete: (taskId: number) => void;
}

/**
 * タスクカードコンポーネント
 * ドラッグ&ドロップ対応、タスクの編集・削除機能付き
 * 新機能：目標時間、実際の時間、完了状態、タイマー開始機能
 */
const TaskCard: React.FC<TaskCardProps> = ({ task, index, onUpdate, onDelete }) => {
  const [isEditing, setIsEditing] = useState(false);
  const [loading, setLoading] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<Types.UpdateTaskRequest>({
    defaultValues: {
      title: task.title,
      description: task.description || '',
      due_date: task.due_date ? format(parseISO(task.due_date), 'yyyy-MM-dd') : '',
      estimated_time: task.estimated_time || undefined,
      is_completed: task.is_completed || false,
    },
  });

  // タスク更新
  const onSubmit = async (data: Types.UpdateTaskRequest) => {
    try {
      setLoading(true);
      const updatedTask = await taskApi.updateTask(task.id, {
        ...data,
        due_date: data.due_date || undefined,
        estimated_time: data.estimated_time || undefined,
      });
      onUpdate(updatedTask);
      setIsEditing(false);
    } catch (error) {
      console.error('タスクの更新に失敗しました:', error);
    } finally {
      setLoading(false);
    }
  };

  // 完了状態の切り替え
  const toggleCompletion = async () => {
    try {
      setLoading(true);
      const updatedTask = await taskApi.updateTask(task.id, {
        is_completed: !task.is_completed,
      });
      onUpdate(updatedTask);
    } catch (error) {
      console.error('タスクの完了状態変更に失敗しました:', error);
    } finally {
      setLoading(false);
    }
  };

  // タイマー開始
  const startTimer = async () => {
    try {
      const duration = (task.estimated_time || 25) * 60; // 分を秒に変換、デフォルト25分
      await timerApi.startTimer(task.id, duration);
      // タイマー開始成功を通知（必要に応じて）
    } catch (error) {
      console.error('タイマー開始に失敗しました:', error);
    }
  };

  // タスク削除
  const handleDelete = async () => {
    if (!window.confirm('このタスクを削除しますか？')) return;
    
    try {
      setLoading(true);
      await taskApi.deleteTask(task.id);
      onDelete(task.id);
    } catch (error) {
      console.error('タスクの削除に失敗しました:', error);
    } finally {
      setLoading(false);
    }
  };

  // 編集モードの切り替え
  const toggleEdit = () => {
    if (isEditing) {
      reset();
    }
    setIsEditing(!isEditing);
  };

  return (
    <Draggable draggableId={`task-${task.id}`} index={index}>
      {(provided, snapshot) => (
        <div
          ref={provided.innerRef}
          {...provided.draggableProps}
          {...provided.dragHandleProps}
          className={`card p-4 mb-3 cursor-move relative ${
            snapshot.isDragging ? 'shadow-lg transform rotate-3' : ''
          } ${task.is_completed ? 'opacity-75' : ''}`}
        >
          {/* 完了時の斜線とDone表示 */}
          {task.is_completed && (
            <>
              {/* 斜線 */}
              <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                <div className="w-full h-0.5 bg-green-500 transform rotate-12"></div>
              </div>
              {/* Done表示 */}
              <div className="absolute top-2 right-2 bg-green-500 text-white text-xs px-2 py-1 rounded font-bold">
                Done
              </div>
            </>
          )}

          {isEditing ? (
            // 編集モード
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
              <div>
                <input
                  type="text"
                  className="input-field text-sm"
                  placeholder="タスク名"
                  {...register('title', { required: 'タスク名は必須です' })}
                />
                {errors.title && (
                  <p className="text-xs text-red-600 mt-1">{errors.title.message}</p>
                )}
              </div>
              
              <div>
                <textarea
                  className="input-field text-sm resize-none"
                  rows={2}
                  placeholder="説明（オプション）"
                  {...register('description')}
                />
              </div>

              <div className="grid grid-cols-2 gap-2">
                <div>
                  <input
                    type="date"
                    className="input-field text-sm"
                    {...register('due_date')}
                  />
                </div>
                <div>
                  <input
                    type="number"
                    min="1"
                    className="input-field text-sm"
                    placeholder="目標時間（分）"
                    {...register('estimated_time', { valueAsNumber: true })}
                  />
                </div>
              </div>

              <div className="flex items-center">
                <input
                  type="checkbox"
                  id={`completed-${task.id}`}
                  className="mr-2"
                  {...register('is_completed')}
                />
                <label htmlFor={`completed-${task.id}`} className="text-sm text-gray-700">
                  完了
                </label>
              </div>
              
              <div className="flex justify-end space-x-2">
                <button
                  type="button"
                  onClick={toggleEdit}
                  className="px-2 py-1 text-xs text-gray-600 hover:text-gray-800"
                >
                  キャンセル
                </button>
                <button
                  type="submit"
                  disabled={loading}
                  className="px-3 py-1 text-xs bg-primary-500 text-white rounded hover:bg-primary-600 disabled:opacity-50"
                >
                  保存
                </button>
              </div>
            </form>
          ) : (
            // 表示モード
            <div>
              <div className="flex justify-between items-start mb-2">
                <h4 className={`text-sm font-medium break-words ${
                  task.is_completed ? 'text-gray-500 line-through' : 'text-gray-900'
                }`}>
                  {task.title}
                </h4>
                <div className="flex space-x-1 ml-2">
                  {/* 完了トグル */}
                  <button
                    onClick={toggleCompletion}
                    disabled={loading}
                    className={`text-sm ${
                      task.is_completed ? 'text-green-500 hover:text-green-600' : 'text-gray-400 hover:text-green-500'
                    }`}
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                  </button>
                  
                  {/* タイマー開始（完了していないタスクのみ） */}
                  {!task.is_completed && (
                    <button
                      onClick={startTimer}
                      className="text-gray-400 hover:text-blue-500"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </button>
                  )}
                  
                  <button
                    onClick={toggleEdit}
                    className="text-gray-400 hover:text-gray-600"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                  </button>
                  <button
                    onClick={handleDelete}
                    disabled={loading}
                    className="text-gray-400 hover:text-red-500 disabled:opacity-50"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </div>
              
              {task.description && (
                <p className={`text-xs mb-2 break-words ${
                  task.is_completed ? 'text-gray-400 line-through' : 'text-gray-600'
                }`}>
                  {task.description}
                </p>
              )}

              {/* 時間情報 */}
              <div className="flex flex-wrap gap-2 mb-2">
                {task.estimated_time && (
                  <span className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-blue-100 text-blue-800">
                    <svg className="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    目標: {task.estimated_time}分
                  </span>
                )}
                {task.actual_time && (
                  <span className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-green-100 text-green-800">
                    <svg className="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    実際: {task.actual_time}分
                  </span>
                )}
              </div>
              
              {task.due_date && (
                <div className={`flex items-center text-xs mb-2 ${
                  task.is_completed ? 'text-gray-400 line-through' : 'text-gray-500'
                }`}>
                  <svg className="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                  {format(parseISO(task.due_date), 'yyyy/MM/dd')}
                </div>
              )}
              
              <div className="mt-2 flex justify-between items-center text-xs text-gray-400">
                <span>#{task.id}</span>
                {loading && (
                  <svg className="animate-spin h-3 w-3" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                )}
              </div>
            </div>
          )}
        </div>
      )}
    </Draggable>
  );
};

export default TaskCard; 