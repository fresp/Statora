import React, { useState } from 'react';

interface UptimeBar {
  date: string;
  uptimePercent: number;
  status: string;
}

interface UptimeTimelineProps {
  bars: UptimeBar[];
  onBarHover?: (bar: UptimeBar | null) => void;
}

const STATUS_COLORS = {
  operational: 'bg-green-500',
  degraded_performance: 'bg-yellow-500', 
  partial_outage: 'bg-orange-500',
  major_outage: 'bg-red-500',
  maintenance: 'bg-purple-500',
};

const STATUS_LABELS = {
  operational: 'Operational',
  degraded_performance: 'Degraded',
  partial_outage: 'Partial Outage', 
  major_outage: 'Major Outage',
  maintenance: 'Maintenance',
};

const getStatusFromPercent = (percent: number): keyof typeof STATUS_COLORS => {
  if (percent === 100) return 'operational';
  if (percent >= 99) return 'degraded_performance';
  if (percent >= 50) return 'partial_outage';
  return 'major_outage';
};

export function UptimeTimeline({ bars, onBarHover }: UptimeTimelineProps) {
  const [hoveredBar, setHoveredBar] = useState<UptimeBar | null>(null);

  // Group bars by month for better visualization
  const groupedBars: Record<string, UptimeBar[]> = {};
  bars.forEach(bar => {
    const month = bar.date.substring(0, 7); // YYYY-MM format
    if (!groupedBars[month]) {
      groupedBars[month] = [];
    }
    groupedBars[month].push(bar);
  });

  const months = Object.keys(groupedBars).sort();

  const handleMouseEnter = (bar: UptimeBar) => {
    setHoveredBar(bar);
    if (onBarHover) {
      onBarHover(bar);
    }
  };

  const handleMouseLeave = () => {
    setHoveredBar(null);
    if (onBarHover) {
      onBarHover(null);
    }
  };

  // Render each month's bars in a single timeline row 
  return (
    <div className="space-y-2">
      {months.map(month => (
        <div key={month} className="flex items-center">
          <span className="text-xs text-gray-500 w-16 mr-2">{month}</span>
          <div className="flex-1 flex gap-px">
            {groupedBars[month].map((bar, idx) => {
              // Determine status from percentage if not provided
              const computedStatus = bar.status || getStatusFromPercent(bar.uptimePercent);
              const status = computedStatus as keyof typeof STATUS_COLORS;
              
              return (
                <div
                  key={`${month}-${idx}`}
                  className={`
                    flex-1 h-4 cursor-pointer border border-gray-200 rounded-sm
                    ${STATUS_COLORS[status]}
                  `}
                  onMouseEnter={() => handleMouseEnter(bar)}
                  onMouseLeave={handleMouseLeave}
                  title={`${bar.date}: ${STATUS_LABELS[status]} (Uptime: ${bar.uptimePercent.toFixed(1)}%)`}
                />
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );
}