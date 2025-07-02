import React, { useState } from 'react';
import { Draggable } from '@hello-pangea/dnd';
import * as Types from '../types';
import { taskApi } from '../lib/api';
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
      order: task.order,
    },
  });

  // タスク更新
  const onSubmit = async (data: Types.UpdateTaskRequest) => {
    try {
      setLoading(true);
      const updatedTask = await taskApi.updateTask(task.id, {
        ...data,
        due_date: data.due_date || undefined,
      });
      onUpdate(updatedTask);
      setIsEditing(false);
    } catch (error) {
      console.error('タスクの更新に失敗しました:', error);
    } finally {
      setLoading(false);
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
    <Draggable draggableId={task.id.toString()} index={index}>
      {(provided, snapshot) => (
        <div
          ref={provided.innerRef}
          {...provided.draggableProps}
          {...provided.dragHandleProps}
          className={`card p-4 mb-3 cursor-move ${
            snapshot.isDragging ? 'shadow-lg transform rotate-3' : ''
          }`}
        >
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
              
              <div>
                <input
                  type="date"
                  className="input-field text-sm"
                  {...register('due_date')}
                />
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
                <h4 className="text-sm font-medium text-gray-900 break-words">{task.title}</h4>
                <div className="flex space-x-1 ml-2">
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
                <p className="text-xs text-gray-600 mb-2 break-words">{task.description}</p>
              )}
              
              {task.due_date && (
                <div className="flex items-center text-xs text-gray-500">
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