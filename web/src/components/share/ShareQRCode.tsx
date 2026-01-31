import { QRCodeSVG } from 'qrcode.react';

export interface ShareQRCodeProps {
  url: string;
  size?: number;
}

export function ShareQRCode({ url, size = 128 }: ShareQRCodeProps) {
  return (
    <div className="bg-white p-3 rounded-lg">
      <QRCodeSVG
        value={url}
        size={size}
        level="M"
        bgColor="#ffffff"
        fgColor="#0a0a0a"
      />
    </div>
  );
}
