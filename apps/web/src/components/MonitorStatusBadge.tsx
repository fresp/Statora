import React from 'react';

interface MonitorStatusBadgeProps {
  status: 'up' | 'down' | 'unknown';
}

export function MonitorStatusBadge({ status }: MonitorStatusBadgeProps) {
  const getStatusConfig = () => {
    switch(status) {
      case 'up':
        return {
          text: 'Operational',
          bgColor: 'bg-green-100',
          textColor: 'text-green-800',
          icon: (
            <svg className="w-3 h-3" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="12" cy="12" r="12" fill="#10B981"/>
            </svg>
          )
        };
      case 'down':
        return {
          text: 'Outage',
          bgColor: 'bg-red-100',
          textColor: 'text-red-800', 
          icon: (
            <svg className="w-3 h-3" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="12" cy="12" r="12" fill="#EF4444"/>
            </svg>
          )
        };
      default:
        return {
          text: 'Unknown',
          bgColor: 'bg-gray-100',
          textColor: 'text-gray-800',
          icon: (
            <svg className="w-3 h-3" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="12" cy="12" r="12" fill="#6B7280"/>
            </svg>
          )
        };
    }
  };

  const config = getStatusConfig();

  return (
    <span className={`
      inline-flex items-center gap-1 px-2 py-1 
      rounded-full text-xs font-medium 
      ${config.bgColor} ${config.textColor}
    `}>
      {config.icon}
      {config.text}
    </span>
  );
}