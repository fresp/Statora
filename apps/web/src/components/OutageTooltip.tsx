import React from 'react';

interface OutageTooltipProps {
  bar: {
    date: string;
    uptimePercent: number;
    status: string;
  } | null;
}

export function OutageTooltip({ bar }: OutageTooltipProps) {
  if (!bar) {
    return null;
  }

  // Map status to human-readable labels
  const getStatusLabel = (status: string) => {
    switch(status) {
      case 'operational':
        return 'Fully Operational';
      case 'degraded_performance':
        return 'Degraded Performance';
      case 'partial_outage':
        return 'Partial Outage';
      case 'major_outage':
        return 'Major Outage';
      case 'maintenance':
        return 'Planned Maintenance';
      default:
        return status.charAt(0).toUpperCase() + status.slice(1);
    }
  };

  // Calculate if there was an outage based on status
  const hasOutage = ['major_outage', 'partial_outage', 'degraded_performance'].includes(bar.status);

  return (
    <div className="absolute z-50 bg-white rounded-md shadow-lg border border-gray-200 p-4 max-w-xs">
      <div className="font-semibold text-gray-900">
        {new Date(bar.date).toLocaleDateString('en-US', { 
          weekday: 'long', 
          year: 'numeric', 
          month: 'long', 
          day: 'numeric' 
        })}
      </div>
      <div className="mt-2 text-sm">
        <div className="flex justify-between">
          <span className="text-gray-600">Status:</span>
          <span className={`font-medium ${
            bar.status === 'operational' ? 'text-green-600' : 
            bar.status === 'major_outage' ? 'text-red-600' : 
            bar.status === 'partial_outage' ? 'text-orange-600' :
            bar.status === 'degraded_performance' ? 'text-yellow-600' :
            'text-purple-600'
          }`}>
            {getStatusLabel(bar.status)}
          </span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-600">Uptime:</span>
          <span className="font-medium">
            {bar.uptimePercent.toFixed(2)}%
          </span>
        </div>
        {hasOutage && (
          <div className="mt-2 pt-2 border-t border-gray-100">
            <div className="text-xs text-gray-500 italic">
              Affected components may have experienced downtime on this date.
            </div>
          </div>
        )}
      </div>
    </div>
  );
}