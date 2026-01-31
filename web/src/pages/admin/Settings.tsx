import { Smartphone, CheckCircle, XCircle, LogOut, RefreshCw } from 'lucide-react';
import { Card, CardTitle, CardDescription, Button, Loading, Spinner } from '@/components/ui';
import { useStatus, useQRCode, useLogoutWhatsApp } from '@/hooks';
import { formatDateTime } from '@/lib/utils';

export default function AdminSettings() {
  const { data: status, isLoading: statusLoading, refetch } = useStatus();
  const { data: qrData, isLoading: qrLoading, refetch: refetchQR } = useQRCode();
  const logoutMutation = useLogoutWhatsApp();

  const isConnected = status?.connected && status?.logged_in;

  if (statusLoading) {
    return (
      <div className="p-8 flex items-center justify-center h-full">
        <Loading message="Loading settings..." />
      </div>
    );
  }

  return (
    <div className="p-8 max-w-2xl">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-text-primary">Settings</h1>
        <p className="text-text-secondary mt-1">
          Manage your WhatsApp connection
        </p>
      </div>

      <Card padding="lg">
        <div className="flex items-start gap-4 mb-6">
          <div className={`p-3 rounded-xl ${isConnected ? 'bg-accent/20' : 'bg-error/20'}`}>
            <Smartphone className={`h-6 w-6 ${isConnected ? 'text-accent' : 'text-error'}`} />
          </div>
          <div>
            <CardTitle>WhatsApp Connection</CardTitle>
            <CardDescription>
              {isConnected
                ? 'Your WhatsApp account is connected and ready'
                : 'Connect your WhatsApp account to enable file uploads'}
            </CardDescription>
          </div>
        </div>

        {isConnected ? (
          /* Connected state */
          <div className="space-y-4">
            <div className="bg-surface-hover rounded-lg p-4 space-y-3">
              <div className="flex items-center gap-2">
                <CheckCircle className="h-4 w-4 text-accent" />
                <span className="text-sm text-text-primary">Connected</span>
              </div>
              
              {status?.phone_number && (
                <div className="flex items-center justify-between">
                  <span className="text-sm text-text-secondary">Phone Number</span>
                  <span className="text-sm text-text-primary">{status.phone_number}</span>
                </div>
              )}
              
              {status?.push_name && (
                <div className="flex items-center justify-between">
                  <span className="text-sm text-text-secondary">Name</span>
                  <span className="text-sm text-text-primary">{status.push_name}</span>
                </div>
              )}
              
              {status?.connected_at && (
                <div className="flex items-center justify-between">
                  <span className="text-sm text-text-secondary">Connected Since</span>
                  <span className="text-sm text-text-primary">
                    {formatDateTime(status.connected_at)}
                  </span>
                </div>
              )}
              
              {status?.reconnect_count !== undefined && status.reconnect_count > 0 && (
                <div className="flex items-center justify-between">
                  <span className="text-sm text-text-secondary">Reconnections</span>
                  <span className="text-sm text-text-primary">{status.reconnect_count}</span>
                </div>
              )}
            </div>

            <Button
              variant="danger"
              className="w-full"
              onClick={() => logoutMutation.mutate()}
              loading={logoutMutation.isPending}
            >
              <LogOut className="h-4 w-4" />
              Disconnect WhatsApp
            </Button>
          </div>
        ) : (
          /* Disconnected state - Show QR code */
          <div className="space-y-4">
            <div className="bg-surface-hover rounded-lg p-4">
              <div className="flex items-center gap-2 mb-4">
                <XCircle className="h-4 w-4 text-error" />
                <span className="text-sm text-text-primary">Not Connected</span>
              </div>
              
              <p className="text-sm text-text-secondary mb-4">
                Scan the QR code below with your WhatsApp mobile app to connect:
              </p>

              {/* QR Code */}
              <div className="flex justify-center py-4">
                {qrLoading ? (
                  <div className="w-48 h-48 flex items-center justify-center">
                    <Spinner size="lg" />
                  </div>
                ) : qrData?.qr_code ? (
                  <div className="bg-white p-4 rounded-lg">
                    <img
                      src={`data:image/png;base64,${qrData.qr_code}`}
                      alt="WhatsApp QR Code"
                      className="w-48 h-48"
                    />
                  </div>
                ) : (
                  <div className="w-48 h-48 bg-surface flex items-center justify-center rounded-lg border border-border">
                    <p className="text-sm text-text-secondary text-center px-4">
                      QR code unavailable
                    </p>
                  </div>
                )}
              </div>

              <p className="text-xs text-text-secondary text-center">
                QR code refreshes automatically every 20 seconds
              </p>
            </div>

            <Button
              variant="secondary"
              className="w-full"
              onClick={() => {
                refetch();
                refetchQR();
              }}
            >
              <RefreshCw className="h-4 w-4" />
              Refresh Status
            </Button>
          </div>
        )}
      </Card>
    </div>
  );
}
