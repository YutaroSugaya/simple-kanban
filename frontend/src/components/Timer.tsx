import React, { useState, useEffect, useCallback } from 'react';
import * as Types from '../types';
import { timerApi } from '../lib/api';

interface TimerProps {
  task?: Types.Task;
  onTimerComplete?: () => void;
  onTimerStop?: (actualDuration: number) => void;
  externalTaskId?: number; // 外部からタイマー開始するタスクID
}

/**
 * 視覚的なタイマーコンポーネント
 * 円形プログレスバーとカウントダウン機能付き
 */
const Timer: React.FC<TimerProps> = ({ task, onTimerComplete, onTimerStop, externalTaskId }) => {
  const [activeSession, setActiveSession] = useState<Types.TimerSession | null>(null);
  const [timeLeft, setTimeLeft] = useState(0);
  const [isRunning, setIsRunning] = useState(false);
  const [customDuration, setCustomDuration] = useState(25); // デフォルト25分
  const [isPaused, setIsPaused] = useState(false);
  const [currentTaskId, setCurrentTaskId] = useState<number | undefined>(undefined);

  // アクティブタイマーの取得
  useEffect(() => {
    (async () => {
      try {
        const timer = await timerApi.getActiveTimer();
        if (timer) {
          setActiveSession(timer);
          setIsRunning(true);
          setIsPaused(false);
          setCurrentTaskId(timer.task_id);
          const elapsed = Math.floor((Date.now() - new Date(timer.start_time).getTime()) / 1000);
          const remaining = Math.max(0, timer.duration - elapsed);
          setTimeLeft(remaining);
        }
      } catch {
        // no-op
      }
    })();
  }, []);

  // 外部からのタイマー開始要求を処理
  useEffect(() => {
    if (externalTaskId && task && task.id === externalTaskId) {
      // startTimer はこの useEffect より後で宣言されるため、直接呼ばずに同等処理を実行
      const run = async () => {
        const durationMinutes = task?.estimated_time || customDuration;
        const duration = durationMinutes * 60;
        const session = await timerApi.startTimer(task.id, duration);
        setActiveSession(session);
        setTimeLeft(duration);
        setIsRunning(true);
        setIsPaused(false);
        setCurrentTaskId(task.id);
        setCustomDuration(durationMinutes);
      };
      run();
    }
  }, [externalTaskId, task, customDuration]);

  // 選択タスクが変わったら記録
  useEffect(() => {
    if (task?.id) {
      setCurrentTaskId(task.id);
    }
  }, [task?.id]);

  // タスクが変更されたときにデフォルト時間を更新
  useEffect(() => {
    if (task?.estimated_time && !isRunning) {
      setCustomDuration(task.estimated_time);
      setTimeLeft(task.estimated_time * 60);
    }
  }, [task, isRunning]);

  // タイマーカウントダウン
  useEffect(() => {
    let interval: number;

    if (isRunning && timeLeft > 0) {
      interval = setInterval(() => {
        setTimeLeft((prev) => {
          if (prev <= 1) {
            // タイマー完了
            setIsRunning(false);
            onTimerComplete?.();
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    }

    return () => clearInterval(interval);
  }, [isRunning, timeLeft, onTimerComplete]);

  // タイマー開始
  const startTimer = useCallback(async () => {
    if (!task && !currentTaskId) return;

    try {
      // タスクの目標時間があれば使用、なければカスタム時間を使用
      const durationMinutes = task?.estimated_time || customDuration;
      const duration = durationMinutes * 60; // 分を秒に変換
      
      const session = await timerApi.startTimer(task ? task.id : (currentTaskId as number), duration);
      setActiveSession(session);
      setTimeLeft(duration);
      setIsRunning(true);
      setIsPaused(false);
      setCurrentTaskId(task ? task.id : currentTaskId);
      
      // カスタム時間も更新
      setCustomDuration(durationMinutes);
    } catch (error) {
      console.error('Failed to start timer:', error);
    }
  }, [task, customDuration, currentTaskId]);

  // タイマー停止
  const pauseTimer = useCallback(async () => {
    if (!activeSession) return;

    try {
      const stoppedSession = await timerApi.stopTimer(activeSession.id!);
      const actualDuration = stoppedSession.duration;
      // timeLeft は減り続けているのでそのまま保持
      setActiveSession(null);
      setIsRunning(false);
      setIsPaused(true);
      onTimerStop?.(actualDuration);
    } catch (error) {
      console.error('Failed to stop timer:', error);
    }
  }, [activeSession, onTimerStop]);

  const resumeTimer = useCallback(async () => {
    if (!currentTaskId || timeLeft <= 0) return;
    try {
      const session = await timerApi.startTimer(currentTaskId, timeLeft);
      setActiveSession(session);
      setIsRunning(true);
      setIsPaused(false);
    } catch (error) {
      console.error('Failed to resume timer:', error);
    }
  }, [currentTaskId, timeLeft]);

  // タイマーリセット
  const resetTimer = useCallback(() => {
    setIsRunning(false);
    setIsPaused(false);
    setActiveSession(null);
    setTimeLeft(customDuration * 60);
  }, [customDuration]);

  // 時間フォーマット関数
  const formatTime = (seconds: number): string => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  // 進捗パーセンテージ
  const progress = activeSession 
    ? Math.max(0, (1 - timeLeft / activeSession.duration) * 100)
    : 0;

  // 円形プログレスバーの計算
  const radius = 60;
  const circumference = 2 * Math.PI * radius;
  const strokeDasharray = circumference;
  const strokeDashoffset = circumference - (progress / 100) * circumference;

  return (
    <div className="text-center">
      {/* タイマー表示 */}
      <div className="relative inline-block">
        {/* 円形プログレスバー */}
        <svg className="w-32 h-32 transform -rotate-90" viewBox="0 0 140 140">
          {/* 背景円 */}
          <circle
            cx="70"
            cy="70"
            r={radius}
            stroke="#e5e7eb"
            strokeWidth="8"
            fill="none"
          />
          {/* プログレス円 */}
          <circle
            cx="70"
            cy="70"
            r={radius}
            stroke="#ef4444"
            strokeWidth="8"
            fill="none"
            strokeLinecap="round"
            strokeDasharray={strokeDasharray}
            strokeDashoffset={strokeDashoffset}
            className="transition-all duration-1000 ease-out"
          />
          {/* プログレス円のグラデーション効果 */}
          <defs>
            <linearGradient id="progressGradient" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" stopColor="#ef4444" />
              <stop offset="100%" stopColor="#f97316" />
            </linearGradient>
          </defs>
          <circle
            cx="70"
            cy="70"
            r={radius}
            stroke="url(#progressGradient)"
            strokeWidth="8"
            fill="none"
            strokeLinecap="round"
            strokeDasharray={strokeDasharray}
            strokeDashoffset={strokeDashoffset}
            className="transition-all duration-1000 ease-out"
            style={{ filter: 'blur(1px)' }}
          />
        </svg>
        
        {/* 中央の時間表示 */}
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="text-center">
            <div className="text-2xl font-bold text-gray-800 font-mono">
              {formatTime(timeLeft)}
            </div>
            {activeSession && (
              <div className="text-xs text-gray-600 mt-1 truncate max-w-24">
                {activeSession.task?.title || 'タイマー実行中'}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* 操作ボタン */}
      <div className="mt-4 flex justify-center space-x-2">
        {!isRunning ? (
          isPaused && timeLeft > 0 ? (
            <button
              onClick={resumeTimer}
              className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 font-medium transition-colors text-sm"
            >
              再開
            </button>
          ) : (
            <button
              onClick={startTimer}
              disabled={!task && !currentTaskId}
              className="px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 disabled:opacity-50 disabled:cursor-not-allowed font-medium transition-colors text-sm"
            >
              開始
            </button>
          )
        ) : (
          <button
            onClick={pauseTimer}
            className="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 font-medium transition-colors text-sm"
          >
            一時停止
          </button>
        )}
        
        <button
          onClick={resetTimer}
          className="px-4 py-2 bg-gray-500 text-white rounded-lg hover:bg-gray-600 font-medium transition-colors text-sm"
        >
          リセット
        </button>
      </div>

      {/* 時間設定 */}
      {!isRunning && (
        <div className="mt-4">
          <label className="block text-sm font-medium text-gray-700 mb-2">
            時間設定（分）
          </label>
          <input
            type="number"
            min="1"
            max="120"
            value={customDuration}
            onChange={(e) => setCustomDuration(parseInt(e.target.value) || 25)}
            className="w-20 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
      )}

      {/* タスク選択メッセージ */}
      {!task && (
        <div className="mt-4 p-3 bg-blue-50 text-blue-700 rounded-lg text-sm">
          タスクを選択してタイマーを開始してください
        </div>
      )}

      {/* 完了時のメッセージ */}
      {timeLeft === 0 && activeSession && (
        <div className="mt-4 p-3 bg-green-50 text-green-700 rounded-lg text-sm">
          <div className="font-medium">タイマー完了！</div>
          <div>お疲れさまでした</div>
        </div>
      )}
    </div>
  );
};

export default Timer; 