import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { calendarApi } from '../lib/api';
import * as Types from '../types';
import { useForm } from 'react-hook-form';

/**
 * 設定ページコンポーネント
 */
const Settings: React.FC = () => {
  const { user, logout } = useAuth();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<Types.UpdateCalendarSettingsRequest>();

  // カレンダー設定の取得
  const fetchSettings = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await calendarApi.getSettings();
      
      // フォームにデフォルト値を設定
      setValue('weekday_start_time', data.weekday_start_time);
      setValue('weekday_end_time', data.weekday_end_time);
      setValue('weekend_start_time', data.weekend_start_time);
      setValue('weekend_end_time', data.weekend_end_time);
      setValue('time_slot_duration', data.time_slot_duration);
    } catch (error: any) {
      setError(error.response?.data?.error || 'カレンダー設定の取得に失敗しました');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSettings();
  }, []);

  // 設定保存
  const onSubmit = async (data: Types.UpdateCalendarSettingsRequest) => {
    try {
      setSaving(true);
      setError(null);
      setSuccessMessage(null);
      
      const updatedSettings = await calendarApi.updateSettings(data);
      setValue('weekday_start_time', updatedSettings.weekday_start_time);
      setValue('weekday_end_time', updatedSettings.weekday_end_time);
      setValue('weekend_start_time', updatedSettings.weekend_start_time);
      setValue('weekend_end_time', updatedSettings.weekend_end_time);
      setValue('time_slot_duration', updatedSettings.time_slot_duration);
      setSuccessMessage('設定を保存しました');
      
      // 成功メッセージを3秒後に消す
      setTimeout(() => setSuccessMessage(null), 3000);
    } catch (error: any) {
      setError(error.response?.data?.error || '設定の保存に失敗しました');
    } finally {
      setSaving(false);
    }
  };

  // 時間選択肢を生成（30分刻み）
  const generateTimeOptions = () => {
    const options = [];
    for (let hour = 0; hour < 24; hour++) {
      for (let minute = 0; minute < 60; minute += 30) {
        const timeString = `${hour.toString().padStart(2, '0')}:${minute.toString().padStart(2, '0')}`;
        options.push(timeString);
      }
    }
    return options;
  };

  const timeOptions = generateTimeOptions();

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary-500"></div>
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
              <h1 className="text-xl font-semibold text-gray-900">設定</h1>
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

      {/* メインコンテンツ */}
      <div className="max-w-4xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-6">カレンダー表示設定</h2>
            
            {error && (
              <div className="mb-4 bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded-lg">
                {error}
              </div>
            )}

            {successMessage && (
              <div className="mb-4 bg-green-50 border border-green-200 text-green-600 px-4 py-3 rounded-lg">
                {successMessage}
              </div>
            )}

            <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
              {/* 平日設定 */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <h3 className="text-md font-medium text-gray-900 mb-4">平日（月〜金）</h3>
                  <div className="space-y-4">
                    <div>
                      <label htmlFor="weekday_start_time" className="block text-sm font-medium text-gray-700 mb-2">
                        開始時刻
                      </label>
                      <select
                        id="weekday_start_time"
                        className="input-field"
                        {...register('weekday_start_time', { required: '開始時刻は必須です' })}
                      >
                        {timeOptions.map((time) => (
                          <option key={time} value={time}>
                            {time}
                          </option>
                        ))}
                      </select>
                      {errors.weekday_start_time && (
                        <p className="mt-1 text-sm text-red-600">{errors.weekday_start_time.message}</p>
                      )}
                    </div>
                    <div>
                      <label htmlFor="weekday_end_time" className="block text-sm font-medium text-gray-700 mb-2">
                        終了時刻
                      </label>
                      <select
                        id="weekday_end_time"
                        className="input-field"
                        {...register('weekday_end_time', { required: '終了時刻は必須です' })}
                      >
                        {timeOptions.map((time) => (
                          <option key={time} value={time}>
                            {time}
                          </option>
                        ))}
                      </select>
                      {errors.weekday_end_time && (
                        <p className="mt-1 text-sm text-red-600">{errors.weekday_end_time.message}</p>
                      )}
                    </div>
                  </div>
                </div>

                {/* 土日設定 */}
                <div>
                  <h3 className="text-md font-medium text-gray-900 mb-4">土日</h3>
                  <div className="space-y-4">
                    <div>
                      <label htmlFor="weekend_start_time" className="block text-sm font-medium text-gray-700 mb-2">
                        開始時刻
                      </label>
                      <select
                        id="weekend_start_time"
                        className="input-field"
                        {...register('weekend_start_time', { required: '開始時刻は必須です' })}
                      >
                        {timeOptions.map((time) => (
                          <option key={time} value={time}>
                            {time}
                          </option>
                        ))}
                      </select>
                      {errors.weekend_start_time && (
                        <p className="mt-1 text-sm text-red-600">{errors.weekend_start_time.message}</p>
                      )}
                    </div>
                    <div>
                      <label htmlFor="weekend_end_time" className="block text-sm font-medium text-gray-700 mb-2">
                        終了時刻
                      </label>
                      <select
                        id="weekend_end_time"
                        className="input-field"
                        {...register('weekend_end_time', { required: '終了時刻は必須です' })}
                      >
                        {timeOptions.map((time) => (
                          <option key={time} value={time}>
                            {time}
                          </option>
                        ))}
                      </select>
                      {errors.weekend_end_time && (
                        <p className="mt-1 text-sm text-red-600">{errors.weekend_end_time.message}</p>
                      )}
                    </div>
                  </div>
                </div>
              </div>

              {/* 時間スロット設定 */}
              <div>
                <label htmlFor="time_slot_duration" className="block text-sm font-medium text-gray-700 mb-2">
                  時間スロット（分）
                </label>
                <select
                  id="time_slot_duration"
                  className="input-field max-w-xs"
                  {...register('time_slot_duration', { 
                    required: '時間スロットは必須です',
                    valueAsNumber: true 
                  })}
                >
                  <option value={10}>10分</option>
                  <option value={15}>15分</option>
                  <option value={30}>30分</option>
                  <option value={60}>60分</option>
                </select>
                {errors.time_slot_duration && (
                  <p className="mt-1 text-sm text-red-600">{errors.time_slot_duration.message}</p>
                )}
                <p className="mt-1 text-sm text-gray-500">
                  カレンダーで表示する最小時間単位を設定します
                </p>
              </div>

              {/* 保存ボタン */}
              <div className="flex justify-end space-x-3">
                <Link
                  to="/"
                  className="btn-secondary"
                >
                  キャンセル
                </Link>
                <button
                  type="submit"
                  disabled={saving}
                  className="btn-primary"
                >
                  {saving ? (
                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                  ) : null}
                  設定を保存
                </button>
              </div>
            </form>

            {/* 設定説明 */}
            <div className="mt-8 p-4 bg-blue-50 rounded-lg">
              <h3 className="text-sm font-medium text-blue-900 mb-2">設定について</h3>
              <ul className="text-sm text-blue-800 space-y-1">
                <li>• 平日と土日で異なる時間設定が可能です</li>
                <li>• 時間スロットはカレンダーの最小表示単位です</li>
                <li>• 設定変更後、カレンダー画面で反映されます</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Settings; 