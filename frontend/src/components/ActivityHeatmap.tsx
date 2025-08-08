import React, { useMemo } from 'react';

interface ActivityData {
  date: string;
  count: number;
}

interface ActivityHeatmapProps {
  data: ActivityData[];
  year?: number;
}

/**
 * GitHub草風のアクティビティヒートマップコンポーネント
 * タスク完了数を色の濃淡で表示
 */
const ActivityHeatmap: React.FC<ActivityHeatmapProps> = ({ data, year = new Date().getFullYear() }) => {
  // 週の開始日を日曜日に設定
  const getWeekStart = (date: Date): Date => {
    const d = new Date(date);
    const day = d.getDay();
    const diff = d.getDate() - day;
    return new Date(d.setDate(diff));
  };

  // 年間のすべての日付を生成
  const generateYearDates = (year: number): Date[] => {
    const dates: Date[] = [];
    const start = new Date(year, 0, 1);
    const end = new Date(year + 1, 0, 1);
    
    for (let d = new Date(start); d < end; d.setDate(d.getDate() + 1)) {
      dates.push(new Date(d));
    }
    
    return dates;
  };

  // データをマップに変換
  const dataMap = useMemo(() => {
    const map = new Map<string, number>();
    data.forEach(item => {
      map.set(item.date, item.count);
    });
    return map;
  }, [data]);

  // 年間の日付を週単位でグループ化
  const weeklyData = useMemo(() => {
    const yearDates = generateYearDates(year);
    const weeks: Date[][] = [];
    let currentWeek: Date[] = [];
    
    // 年の最初の日が週の途中から始まる場合の調整
    const firstDate = yearDates[0];
    const firstWeekStart = getWeekStart(firstDate);
    
    // 必要に応じて前年の日付で週を埋める
    if (firstWeekStart < firstDate) {
      for (let d = new Date(firstWeekStart); d < firstDate; d.setDate(d.getDate() + 1)) {
        currentWeek.push(new Date(d));
      }
    }
    
    yearDates.forEach(date => {
      currentWeek.push(date);
      
      if (currentWeek.length === 7) {
        weeks.push(currentWeek);
        currentWeek = [];
      }
    });
    
    // 最後の週が7日未満の場合、次年の日付で埋める
    if (currentWeek.length > 0) {
      const lastDate = currentWeek[currentWeek.length - 1];
      for (let i = currentWeek.length; i < 7; i++) {
        const nextDate = new Date(lastDate);
        nextDate.setDate(lastDate.getDate() + (i - currentWeek.length + 1));
        currentWeek.push(nextDate);
      }
      weeks.push(currentWeek);
    }
    
    return weeks;
  }, [year]);

  // 色の強度を計算
  const getColorIntensity = (count: number): string => {
    if (count === 0) return 'bg-gray-100';
    if (count <= 1) return 'bg-green-100';
    if (count <= 3) return 'bg-green-200';
    if (count <= 5) return 'bg-green-300';
    if (count <= 8) return 'bg-green-400';
    return 'bg-green-500';
  };

  // 月ラベルを生成
  const monthLabels = useMemo(() => {
    const labels: { month: string; startWeek: number }[] = [];
    let currentMonth = -1;
    
    weeklyData.forEach((week, weekIndex) => {
      const midWeekDate = week[3]; // 週の真ん中の日
      if (midWeekDate.getFullYear() === year) {
        const month = midWeekDate.getMonth();
        if (month !== currentMonth) {
          currentMonth = month;
          labels.push({
            month: midWeekDate.toLocaleDateString('ja-JP', { month: 'short' }),
            startWeek: weekIndex
          });
        }
      }
    });
    
    return labels;
  }, [weeklyData, year]);

  // 統計計算
  const stats = useMemo(() => {
    const totalDays = data.length;
    const activeDays = data.filter(d => d.count > 0).length;
    const totalTasks = data.reduce((sum, d) => sum + d.count, 0);
    const maxTasks = Math.max(...data.map(d => d.count), 0);
    
    return { totalDays, activeDays, totalTasks, maxTasks };
  }, [data]);

  return (
    <div className="bg-white rounded-lg shadow-sm border p-6">
      <div className="mb-4">
        <h3 className="text-lg font-medium text-gray-900 mb-2">
          {year}年のタスク完了アクティビティ
        </h3>
        <div className="flex flex-wrap gap-4 text-sm text-gray-600">
          <span>総タスク完了数: <strong>{stats.totalTasks}</strong></span>
          <span>アクティブ日数: <strong>{stats.activeDays}</strong></span>
          <span>最大日完了数: <strong>{stats.maxTasks}</strong></span>
        </div>
      </div>

      {/* ヒートマップ */}
      <div className="overflow-x-auto">
        <div className="inline-block min-w-full">
          {/* 月ラベル */}
          <div className="flex mb-2" style={{ marginLeft: '30px' }}>
            {monthLabels.map((label, index) => (
              <div
                key={index}
                className="text-xs text-gray-500"
                style={{
                  marginLeft: index === 0 ? 0 : `${(label.startWeek - (monthLabels[index - 1]?.startWeek || 0)) * 12}px`,
                  minWidth: '24px'
                }}
              >
                {label.month}
              </div>
            ))}
          </div>

          {/* グリッド */}
          <div className="flex">
            {/* 曜日ラベル */}
            <div className="flex flex-col mr-2">
              {['', '月', '', '水', '', '金', ''].map((day, index) => (
                <div key={index} className="h-3 flex items-center text-xs text-gray-500 mb-1">
                  {day}
                </div>
              ))}
            </div>

            {/* ヒートマップグリッド */}
            <div className="flex gap-1">
              {weeklyData.map((week, weekIndex) => (
                <div key={weekIndex} className="flex flex-col gap-1">
                  {week.map((date, dayIndex) => {
                    const dateStr = date.toISOString().split('T')[0];
                    const count = dataMap.get(dateStr) || 0;
                    const isCurrentYear = date.getFullYear() === year;
                    
                    return (
                      <div
                        key={`${weekIndex}-${dayIndex}`}
                        className={`w-3 h-3 rounded-sm cursor-pointer transition-all duration-200 hover:ring-2 hover:ring-gray-400 ${
                          isCurrentYear ? getColorIntensity(count) : 'bg-gray-50'
                        }`}
                        title={`${date.toLocaleDateString('ja-JP')}: ${count}タスク完了`}
                      />
                    );
                  })}
                </div>
              ))}
            </div>
          </div>

          {/* 凡例 */}
          <div className="flex items-center justify-between mt-4">
            <div className="flex items-center gap-2 text-xs text-gray-600">
              <span>少ない</span>
              <div className="flex gap-1">
                <div className="w-3 h-3 rounded-sm bg-gray-100"></div>
                <div className="w-3 h-3 rounded-sm bg-green-100"></div>
                <div className="w-3 h-3 rounded-sm bg-green-200"></div>
                <div className="w-3 h-3 rounded-sm bg-green-300"></div>
                <div className="w-3 h-3 rounded-sm bg-green-400"></div>
                <div className="w-3 h-3 rounded-sm bg-green-500"></div>
              </div>
              <span>多い</span>
            </div>
            
            {/* GitHub風の説明 */}
            <div className="text-xs text-gray-500">
              GitHubのコミット草をイメージした表示
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ActivityHeatmap; 