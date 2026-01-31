import { 
  Upload, 
  Download, 
  HardDrive, 
  Activity
} from 'lucide-react';
import { Link } from '@tanstack/react-router';
import { 
  AreaChart, 
  Area, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer 
} from 'recharts';
import { Card, CardTitle, Badge, Button, Loading } from '@/components/ui';
import { useStats, useHourlyStats, useFiles } from '@/hooks';
import { formatBytes, formatRelativeTime } from '@/lib/utils';

export default function AdminDashboard() {
  const { data: stats, isLoading: statsLoading } = useStats();
  const { data: hourlyStats, isLoading: chartLoading } = useHourlyStats(24);
  const { data: filesData } = useFiles({ page: 1, per_page: 5 });

  if (statsLoading) {
    return (
      <div className="p-8 flex items-center justify-center h-full">
        <Loading message="Loading dashboard..." />
      </div>
    );
  }

  const statCards = [
    {
      label: 'Total Files',
      value: stats?.storage.active_files || 0,
      icon: HardDrive,
      color: 'text-blue-400',
      bgColor: 'bg-blue-400/20',
    },
    {
      label: 'Total Uploads',
      value: stats?.realtime.uploads_total || 0,
      icon: Upload,
      color: 'text-accent',
      bgColor: 'bg-accent/20',
    },
    {
      label: 'Total Downloads',
      value: stats?.realtime.downloads_total || 0,
      icon: Download,
      color: 'text-purple-400',
      bgColor: 'bg-purple-400/20',
    },
    {
      label: 'Storage Used',
      value: formatBytes(stats?.storage.total_bytes || 0),
      icon: Activity,
      color: 'text-orange-400',
      bgColor: 'bg-orange-400/20',
    },
  ];

  // Transform hourly stats for chart
  const chartData = hourlyStats?.map((item) => ({
    hour: new Date(item.hour).toLocaleTimeString('en-US', { hour: 'numeric' }),
    uploads: item.uploads,
    downloads: item.downloads,
  })) || [];

  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-text-primary">Dashboard</h1>
        <p className="text-text-secondary mt-1">
          Overview of your WhatsBox service
        </p>
      </div>

      {/* Stat Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        {statCards.map((stat) => (
          <Card key={stat.label} padding="md">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-sm text-text-secondary">{stat.label}</p>
                <p className="text-2xl font-bold text-text-primary mt-1">
                  {stat.value}
                </p>
              </div>
              <div className={`p-2 rounded-lg ${stat.bgColor}`}>
                <stat.icon className={`h-5 w-5 ${stat.color}`} />
              </div>
            </div>
          </Card>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Activity Chart */}
        <div className="lg:col-span-2">
          <Card padding="md">
            <div className="flex items-center justify-between mb-6">
              <CardTitle>Activity (24h)</CardTitle>
              <div className="flex items-center gap-4 text-sm">
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 rounded-full bg-accent" />
                  <span className="text-text-secondary">Uploads</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 rounded-full bg-purple-400" />
                  <span className="text-text-secondary">Downloads</span>
                </div>
              </div>
            </div>
            
            {chartLoading ? (
              <div className="h-64 flex items-center justify-center">
                <Loading />
              </div>
            ) : (
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={chartData}>
                    <defs>
                      <linearGradient id="uploadGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#22c55e" stopOpacity={0.3} />
                        <stop offset="95%" stopColor="#22c55e" stopOpacity={0} />
                      </linearGradient>
                      <linearGradient id="downloadGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#a855f7" stopOpacity={0.3} />
                        <stop offset="95%" stopColor="#a855f7" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="#262626" />
                    <XAxis 
                      dataKey="hour" 
                      stroke="#a1a1aa" 
                      fontSize={12}
                      tickLine={false}
                    />
                    <YAxis 
                      stroke="#a1a1aa" 
                      fontSize={12}
                      tickLine={false}
                      axisLine={false}
                    />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: '#141414',
                        border: '1px solid #262626',
                        borderRadius: '8px',
                      }}
                      labelStyle={{ color: '#fafafa' }}
                    />
                    <Area
                      type="monotone"
                      dataKey="uploads"
                      stroke="#22c55e"
                      fillOpacity={1}
                      fill="url(#uploadGradient)"
                    />
                    <Area
                      type="monotone"
                      dataKey="downloads"
                      stroke="#a855f7"
                      fillOpacity={1}
                      fill="url(#downloadGradient)"
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            )}
          </Card>
        </div>

        {/* Recent Activity */}
        <div>
          <Card padding="md">
            <div className="flex items-center justify-between mb-4">
              <CardTitle>Recent Files</CardTitle>
              <Link to="/admin/files">
                <Button variant="ghost" size="sm">
                  View all
                </Button>
              </Link>
            </div>
            
            <div className="space-y-3">
              {filesData?.files.slice(0, 5).map((file) => (
                <div
                  key={file.id}
                  className="flex items-center justify-between py-2 border-b border-border last:border-0"
                >
                  <div className="min-w-0 flex-1">
                    <p className="text-sm text-text-primary truncate">
                      {file.filename}
                    </p>
                    <p className="text-xs text-text-secondary">
                      {formatBytes(file.file_size)} • {formatRelativeTime(file.created_at)}
                    </p>
                  </div>
                  <Badge
                    variant={file.status === 'active' ? 'success' : 'warning'}
                    size="sm"
                  >
                    {file.status}
                  </Badge>
                </div>
              ))}
              
              {(!filesData?.files || filesData.files.length === 0) && (
                <p className="text-sm text-text-secondary text-center py-4">
                  No files uploaded yet
                </p>
              )}
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
}
