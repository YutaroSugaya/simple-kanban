import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { boardApi, analyticsApi } from '../lib/api';
import * as Types from '../types';
import { useForm } from 'react-hook-form';
import ActivityHeatmap from '../components/ActivityHeatmap';

/**
 * ダッシュボード（ボード一覧）ページコンポーネント
 */
const Dashboard: React.FC = () => {
  const { user, logout } = useAuth();
  const [boards, setBoards] = useState<Types.Board[]>([]);
  const [activityData, setActivityData] = useState<{ date: string; count: number }[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [creating, setCreating] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<Types.CreateBoardRequest>();

  // ボード一覧とアクティビティデータの取得
  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);
      console.log('Fetching boards and activity data...'); // デバッグログ
      
      const [boardsData, activityStatsData] = await Promise.all([
        boardApi.getBoards(),
        analyticsApi.getTaskCompletionStats()
      ]);
      
      console.log('Boards data received:', boardsData); // デバッグログ
      console.log('Activity data received:', activityStatsData); // デバッグログ
      
      // データが配列であることを確認
      if (Array.isArray(boardsData)) {
        setBoards(boardsData);
      } else {
        console.error('Boards data is not an array:', boardsData);
        setBoards([]);
        setError('ボードデータの形式が正しくありません');
      }

      if (Array.isArray(activityStatsData)) {
        setActivityData(activityStatsData);
      } else {
        console.error('Activity data is not an array:', activityStatsData);
        setActivityData([]);
      }
    } catch (error: any) {
      console.error('Error fetching data:', error); // デバッグログ
      setError(error.response?.data?.error || 'データの取得に失敗しました');
      setBoards([]); // エラーが発生した場合も空配列に設定
      setActivityData([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  // ボード作成
  const onSubmit = async (data: Types.CreateBoardRequest) => {
    try {
      setCreating(true);
      console.log('Creating board with data:', data); // デバッグログ
      const newBoard = await boardApi.createBoard(data);
      console.log('New board received:', newBoard); // デバッグログ
      
      // newBoardが正しいデータ構造を持っているかチェック
      if (newBoard && typeof newBoard === 'object' && 'id' in newBoard && 'name' in newBoard) {
        setBoards(prevBoards => {
          const updatedBoards = [...(prevBoards || []), newBoard];
          console.log('Updated boards:', updatedBoards); // デバッグログ
          return updatedBoards;
        });
        setShowCreateModal(false);
        reset();
      } else {
        console.error('Invalid board data received:', newBoard);
        setError('ボードの作成に失敗しました：無効なデータが返されました');
      }
    } catch (error: any) {
      console.error('Board creation error:', error); // デバッグログ
      setError(error.response?.data?.error || 'ボードの作成に失敗しました');
    } finally {
      setCreating(false);
    }
  };

  // ボード削除
  const deleteBoard = async (boardId: number) => {
    if (!window.confirm('このボードを削除しますか？')) return;

    try {
      await boardApi.deleteBoard(boardId);
      setBoards(prevBoards => (prevBoards || []).filter(board => board.id !== boardId));
    } catch (error: any) {
      setError(error.response?.data?.error || 'ボードの削除に失敗しました');
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary-500"></div>
      </div>
    );
  }

  // boardsが存在することを確認
  const safeBoards = boards || [];

  return (
    <div className="min-h-screen bg-gray-50">
      {/* ヘッダー */}
      <nav className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <h1 className="text-xl font-semibold text-gray-900">Simple Kanban</h1>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-600">
                こんにちは、{user?.email}さん
              </span>
              <Link
                to="/calendar"
                className="text-sm text-gray-600 hover:text-gray-800"
              >
                カレンダー
              </Link>
              <Link
                to="/settings"
                className="text-sm text-gray-600 hover:text-gray-800"
              >
                設定
              </Link>
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
          {/* アクティビティヒートマップ */}
          <div className="mb-8">
            <ActivityHeatmap data={activityData} />
          </div>

          <div className="flex justify-between items-center mb-6">
            <h2 className="text-2xl font-bold text-gray-900">マイボード</h2>
            <button
              onClick={() => setShowCreateModal(true)}
              className="btn-primary"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              新しいボード
            </button>
          </div>

          {error && (
            <div className="mb-4 bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded-lg">
              {error}
            </div>
          )}

          {/* ボード一覧 */}
          {safeBoards.length === 0 ? (
            <div className="text-center py-12">
              <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
              <h3 className="mt-2 text-sm font-medium text-gray-900">ボードがありません</h3>
              <p className="mt-1 text-sm text-gray-500">
                新しいボードを作成してKanbanを始めましょう
              </p>
              <div className="mt-6">
                <button
                  onClick={() => setShowCreateModal(true)}
                  className="btn-primary"
                >
                  最初のボードを作成
                </button>
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {safeBoards.map((board) => (
                <div key={board.id} className="card p-6">
                  <div className="flex justify-between items-start mb-4">
                    <h3 className="text-lg font-medium text-gray-900">{board.name}</h3>
                    <button
                      onClick={() => deleteBoard(board.id)}
                      className="text-gray-400 hover:text-red-500"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </div>
                  <p className="text-sm text-gray-500 mb-4">
                    作成日: {new Date(board.created_at).toLocaleDateString('ja-JP')}
                  </p>
                  <Link
                    to={`/boards/${board.id}`}
                    className="w-full btn-primary text-center block"
                  >
                    ボードを開く
                  </Link>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* ボード作成モーダル */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
          <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
            <div className="mt-3">
              <h3 className="text-lg font-medium text-gray-900 mb-4">新しいボード</h3>
              <form onSubmit={handleSubmit(onSubmit)}>
                <div className="mb-4">
                  <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-2">
                    ボード名
                  </label>
                  <input
                    id="name"
                    type="text"
                    className="input-field"
                    placeholder="ボード名を入力"
                    {...register('name', {
                      required: 'ボード名は必須です',
                      minLength: {
                        value: 1,
                        message: 'ボード名は1文字以上で入力してください',
                      },
                    })}
                  />
                  {errors.name && (
                    <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>
                  )}
                </div>
                <div className="flex justify-end space-x-3">
                  <button
                    type="button"
                    onClick={() => {
                      setShowCreateModal(false);
                      reset();
                    }}
                    className="btn-secondary"
                  >
                    キャンセル
                  </button>
                  <button
                    type="submit"
                    disabled={creating}
                    className="btn-primary"
                  >
                    {creating ? (
                      <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                    ) : null}
                    作成
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Dashboard; 