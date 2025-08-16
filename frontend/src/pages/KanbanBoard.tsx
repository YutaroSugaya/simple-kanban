import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { DragDropContext } from '@hello-pangea/dnd';
import { useAuth } from '../contexts/AuthContext';
import { boardApi, taskApi } from '../lib/api';
import * as Types from '../types';
import Column from '../components/Column';

/**
 * Kanbanボードメインページコンポーネント
 * ドラッグ&ドロップによるタスク管理機能付き
 */
const KanbanBoard: React.FC = () => {
  const { boardId } = useParams<{ boardId: string }>();
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  
  const [board, setBoard] = useState<Types.Board | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // boardIdが存在しない場合の早期リターン
  if (!boardId) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded-lg mb-4">
            ボードIDが指定されていません
          </div>
          <Link to="/" className="btn-primary">
            ダッシュボードに戻る
          </Link>
        </div>
      </div>
    );
  }

  // ボード詳細の取得
  const fetchBoard = async () => {
    if (!boardId) {
      setError('ボードIDが指定されていません');
      setLoading(false);
      return;
    }
    
    const parsedBoardId = parseInt(boardId);
    if (isNaN(parsedBoardId)) {
      setError('無効なボードIDです');
      setLoading(false);
      return;
    }
    
    try {
      setLoading(true);
      setError(null);
      const data = await boardApi.getBoardById(parsedBoardId);
      setBoard(data);
    } catch (error: any) {
      setError(error.response?.data?.error || 'ボードの取得に失敗しました');
      if (error.response?.status === 404) {
        setTimeout(() => navigate('/'), 2000);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchBoard();
  }, [boardId]);

  // タスクのドラッグ&ドロップ処理
  const onDragEnd = async (result: any) => {
    const { destination, source, draggableId } = result;

    // ドロップ先がない場合は何もしない
    if (!destination) return;

    // 同じ位置にドロップした場合は何もしない
    if (
      destination.droppableId === source.droppableId &&
      destination.index === source.index
    ) {
      return;
    }

    const taskId = parseInt(draggableId.replace('task-', ''));
    const sourceColumnId = parseInt(source.droppableId);
    const destinationColumnId = parseInt(destination.droppableId);

    // 楽観的UI更新のため、ローカル状態を先に更新
    if (board) {
      const newBoard = { ...board };
      const sourceColumn = newBoard.columns?.find(col => col.id === sourceColumnId);
      const destinationColumn = newBoard.columns?.find(col => col.id === destinationColumnId);

      if (sourceColumn && destinationColumn) {
        // タスクを移動元から削除
        const taskToMove = sourceColumn.tasks?.find(task => task.id === taskId);
        if (taskToMove) {
          sourceColumn.tasks = sourceColumn.tasks?.filter(task => task.id !== taskId) || [];

          // タスクを移動先に追加
          if (!destinationColumn.tasks) destinationColumn.tasks = [];
          destinationColumn.tasks.splice(destination.index, 0, {
            ...taskToMove,
            column_id: destinationColumnId,
            order: destination.index + 1,
          });

          // 順序を更新（1ベースに変換）
          destinationColumn.tasks.forEach((task, index) => {
            task.order = index + 1;
          });

          setBoard(newBoard);
        }
      }
    }

    // APIに変更を送信
    try {
      const moveData = {
        new_column_id: destinationColumnId,
        new_order: destination.index + 1, // 0ベースから1ベースに変換
      };
      await taskApi.moveTask(taskId, moveData);
    } catch (error) {
      
      // エラーが発生した場合、データを再取得
      fetchBoard();
    }
  };

  // タスク更新時の処理
  const handleTaskUpdate = async (updatedTask: Types.Task) => {
    if (!board) return;
    const newBoard = { ...board };

    // 現在の所属カラムを特定
    let currentColumn = newBoard.columns?.find(col => col.id === updatedTask.column_id)
      || newBoard.columns?.find(col => col.tasks?.some(t => t.id === updatedTask.id));

    // 通常の更新反映
    if (currentColumn && currentColumn.tasks) {
      const taskIndex = currentColumn.tasks.findIndex(task => task.id === updatedTask.id);
      if (taskIndex !== -1) {
        currentColumn.tasks[taskIndex] = { ...currentColumn.tasks[taskIndex], ...updatedTask } as Types.Task;
      }
    }

    // 完了 → Done へ移動
    if (updatedTask.is_completed) {
      const doneColumn = newBoard.columns?.find(col => col.title.toLowerCase() === 'done');
      if (doneColumn && currentColumn && currentColumn.id !== doneColumn.id) {
        // 楽観的UI更新
        if (currentColumn?.tasks) {
          currentColumn.tasks = currentColumn.tasks.filter(t => t.id !== updatedTask.id);
        }
        if (!doneColumn.tasks) doneColumn.tasks = [];
        const newOrder = (doneColumn.tasks?.length || 0) + 1;
        doneColumn.tasks.push({ ...updatedTask, column_id: doneColumn.id, order: newOrder } as Types.Task);
        setBoard(newBoard);

        // サーバーに移動反映
        try {
          await taskApi.moveTask(updatedTask.id, { new_column_id: doneColumn.id, new_order: newOrder });
        } catch {
          // 失敗時は再取得で整合
          fetchBoard();
        }
      }
    } else {
      setBoard(newBoard);
    }
  };

  // タスク削除時の処理
  const handleTaskDelete = (taskId: number) => {
    if (board) {
      const newBoard = { ...board };
      newBoard.columns?.forEach(column => {
        if (column.tasks) {
          column.tasks = column.tasks.filter(task => task.id !== taskId);
        }
      });
      setBoard(newBoard);
    }
  };

  // タスク作成時の処理
  const handleTaskCreate = (newTask: Types.Task) => {
    console.log('handleTaskCreate called with:', newTask);
    if (board) {
      const newBoard = { ...board };
      const column = newBoard.columns?.find(col => col.id === newTask.column_id);
      console.log('Found column:', column?.id, 'for task column_id:', newTask.column_id);
      if (column) {
        if (!column.tasks) column.tasks = [];
        column.tasks.push(newTask);
        console.log('Added task to column. New tasks count:', column.tasks.length);
        setBoard(newBoard);
      } else {
        console.error('Column not found for task column_id:', newTask.column_id);
      }
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary-500"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded-lg mb-4">
            {error}
          </div>
          <Link to="/" className="btn-primary">
            ダッシュボードに戻る
          </Link>
        </div>
      </div>
    );
  }

  if (!board) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="bg-yellow-50 border border-yellow-200 text-yellow-800 px-4 py-3 rounded-lg mb-4">
            ボードデータが見つかりません
          </div>
          <Link to="/" className="btn-primary">
            ダッシュボードに戻る
          </Link>
        </div>
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
              <h1 className="text-xl font-semibold text-gray-900">{board.name}</h1>
            </div>
            <div className="flex items-center space-x-4">
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

      {/* Kanbanボード */}
      <div className="p-6">
        <DragDropContext onDragEnd={onDragEnd}>
          <div className="flex space-x-6 overflow-x-auto pb-6">
            {board.columns && board.columns.length > 0 ? (
              <>
                {board.columns
                  .filter((column) => column.title.toLowerCase() !== 'done')
                  .sort((a, b) => a.order - b.order)
                  .map((column) => (
                    <Column
                      key={column.id}
                      column={column}
                      onTaskUpdate={handleTaskUpdate}
                      onTaskDelete={handleTaskDelete}
                      onTaskCreate={handleTaskCreate}
                    />
                  ))}
                {/* Done カラムを最後尾に表示 */}
                {board.columns
                  .filter((column) => column.title.toLowerCase() === 'done')
                  .map((column) => (
                    <Column
                      key={column.id}
                      column={column}
                      onTaskUpdate={handleTaskUpdate}
                      onTaskDelete={handleTaskDelete}
                      onTaskCreate={handleTaskCreate}
                    />
                  ))}
              </>
            ) : (
              <div className="flex-1 flex items-center justify-center py-12">
                <div className="text-center text-gray-500">
                  <svg className="mx-auto h-12 w-12 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                  <p>このボードにはまだカラムがありません</p>
                  <p className="text-sm mt-1">新しいタスクを追加するとカラムが自動作成されます</p>
                </div>
              </div>
            )}
          </div>
        </DragDropContext>
      </div>

      {/* 使用方法のヒント */}
      <div className="fixed bottom-4 right-4">
        <div className="bg-primary-500 text-white p-3 rounded-lg shadow-lg max-w-xs">
          <p className="text-sm">
            💡 タスクをドラッグ&ドロップして移動できます
          </p>
        </div>
      </div>
    </div>
  );
};

export default KanbanBoard; 