import React, { useState } from 'react';
import { Droppable } from '@hello-pangea/dnd';
import * as Types from '../types';
import TaskCard from './TaskCard';
import { taskApi } from '../lib/api';
import { useForm } from 'react-hook-form';

interface ColumnProps {
  column: Types.Column;
  onTaskUpdate: (updatedTask: Types.Task) => void;
  onTaskDelete: (taskId: number) => void;
  onTaskCreate: (newTask: Types.Task) => void;
}

/**
 * カラムコンポーネント
 * タスクのリスト表示とドラッグ&ドロップ対応
 */
const Column: React.FC<ColumnProps> = ({ column, onTaskUpdate, onTaskDelete, onTaskCreate }) => {
  const [showAddForm, setShowAddForm] = useState(false);
  const [creating, setCreating] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<Types.CreateTaskRequest>();

  // タスク作成
  const onSubmit = async (data: Types.CreateTaskRequest) => {
    try {
      setCreating(true);
      const maxOrder = Math.max(...(column.tasks?.map(t => t.order) || [0]), 0);
      const newTask = await taskApi.createTask({
        ...data,
        column_id: column.id,
        order: maxOrder + 1,
        due_date: data.due_date || undefined,
      });
      
      // フロントエンド側でcolumn_idを追加（サーバーから返されないため）
      const taskWithColumnId = {
        ...newTask,
        column_id: column.id
      };
      
      onTaskCreate(taskWithColumnId);
      setShowAddForm(false);
      reset();
    } catch (error) {
      console.error('タスクの作成に失敗しました:', error);
    } finally {
      setCreating(false);
    }
  };

  const cancelAdd = () => {
    setShowAddForm(false);
    reset();
  };

  return (
    <div className="bg-gray-100 rounded-lg p-4 w-80 flex-shrink-0">
      {/* カラムヘッダー */}
      <div className="flex justify-between items-center mb-4">
        <div className="flex items-center">
          <h3 className="font-semibold text-gray-900">{column.title}</h3>
          <span className="ml-2 text-sm text-gray-500 bg-gray-200 px-2 py-1 rounded-full">
            {column.tasks?.length || 0}
          </span>
        </div>
        <button
          onClick={() => setShowAddForm(true)}
          className="text-gray-500 hover:text-gray-700"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
        </button>
      </div>

      {/* タスク追加フォーム */}
      {showAddForm && (
        <div className="card p-3 mb-3">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
            <div>
              <input
                type="text"
                className="input-field text-sm"
                placeholder="タスク名を入力"
                autoFocus
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
                onClick={cancelAdd}
                className="px-3 py-1 text-sm text-gray-600 hover:text-gray-800"
              >
                キャンセル
              </button>
              <button
                type="submit"
                disabled={creating}
                className="px-3 py-1 text-sm bg-primary-500 text-white rounded hover:bg-primary-600 disabled:opacity-50"
              >
                {creating ? (
                  <svg className="animate-spin -ml-1 mr-1 h-3 w-3 text-white" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                ) : null}
                追加
              </button>
            </div>
          </form>
        </div>
      )}

      {/* タスクリスト */}
      <Droppable droppableId={column.id.toString()}>
        {(provided, snapshot) => (
          <div
            ref={provided.innerRef}
            {...provided.droppableProps}
            className={`min-h-[200px] transition-colors ${
              snapshot.isDraggingOver ? 'bg-blue-50 border-2 border-dashed border-blue-300' : ''
            }`}
          >
            {column.tasks && column.tasks.length > 0 ? (
              column.tasks
                .sort((a, b) => a.order - b.order)
                .map((task, index) => (
                  <TaskCard
                    key={task.id}
                    task={task}
                    index={index}
                    onUpdate={onTaskUpdate}
                    onDelete={onTaskDelete}
                  />
                ))
            ) : (
              !showAddForm && (
                <div className="text-center py-8 text-gray-500">
                  <svg className="mx-auto h-8 w-8 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                  <p className="text-sm">タスクがありません</p>
                  <button
                    onClick={() => setShowAddForm(true)}
                    className="mt-2 text-sm text-primary-600 hover:text-primary-800"
                  >
                    最初のタスクを追加
                  </button>
                </div>
              )
            )}
            {provided.placeholder}
          </div>
        )}
      </Droppable>
    </div>
  );
};

export default Column; 