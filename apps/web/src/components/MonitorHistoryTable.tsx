import React from 'react';

interface Outage {
  id: string;
  monitorId: string;
  monitorName: string;  // Not in model but useful to have
  startedAt: string;
  endedAt: string;
  durationSeconds: number;
  status: 'active' | 'resolved';
  componentId?: string;
  componentName?: string;  // Not in model but useful for display
  subComponentId?: string;
  subComponentName?: string;
}

interface MonitorHistoryTableProps {
  outages: Outage[];
}

export function MonitorHistoryTable({ outages }: MonitorHistoryTableProps) {
  // Helper function to format duration in seconds to human readable format
  const formatDuration = (seconds: number) => {
    if (seconds < 0) return "--:--"; // Handle unresolved outages
    
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = Math.floor(seconds % 60);

    if (h > 0) return `${h}h ${m}m ${s}s`;
    if (m > 0) return `${m}m ${s}s`;
    return `${s}s`;
  };

  // Helper to format date strings nicely
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  // Function to get affected component/subcomponent name
  const getAffectedResource = (outage: Outage) => {
    if (outage.componentName) {
      return outage.componentName;
    }
    if (outage.subComponentName) {
      return `${outage.subComponentName} (${outage.componentName || 'parent'})`;
    }
    return 'Unknown';
  };

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full text-sm">
        <thead className="bg-gray-50 border-b border-gray-100">
          <tr>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Monitor Name</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Affected Resource</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Start Time</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">End Time</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Duration</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Status</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-50">
          {outages.map(outage => (
            <tr key={outage.id} className="hover:bg-gray-50">
              <td className="px-4 py-3 font-medium text-gray-900">{outage.monitorName}</td>
              <td className="px-4 py-3 text-gray-500">{getAffectedResource(outage)}</td>
              <td className="px-4 py-3 text-gray-500">{formatDate(outage.startedAt)}</td>
              <td className="px-4 py-3 text-gray-500">
                {outage.endedAt ? formatDate(outage.endedAt) : '--:--'}
              </td>
              <td className="px-4 py-3 text-gray-500">{formatDuration(outage.durationSeconds)}</td>
              <td className="px-4 py-3">
                <span className={`
                  px-2 py-1 text-xs rounded-full
                  ${outage.status === 'active' 
                    ? 'bg-red-100 text-red-800' 
                    : 'bg-green-100 text-green-800'}
                `}>
                  {outage.status.charAt(0).toUpperCase() + outage.status.slice(1)}
                </span>
              </td>
            </tr>
          ))}
          
          {outages.length === 0 && (
            <tr>
              <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                No outage history recorded yet.
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}