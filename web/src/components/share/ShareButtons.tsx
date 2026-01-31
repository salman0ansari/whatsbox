import { Twitter, Facebook, MessageCircle, Send } from 'lucide-react';

export interface ShareButtonsProps {
  url: string;
  title?: string;
}

export function ShareButtons({ url, title = 'Check out this file' }: ShareButtonsProps) {
  const encodedUrl = encodeURIComponent(url);
  const encodedTitle = encodeURIComponent(title);

  const shareLinks = [
    {
      name: 'Twitter',
      icon: Twitter,
      href: `https://twitter.com/intent/tweet?url=${encodedUrl}&text=${encodedTitle}`,
      color: 'hover:bg-[#1DA1F2]/20 hover:text-[#1DA1F2]',
    },
    {
      name: 'Facebook',
      icon: Facebook,
      href: `https://www.facebook.com/sharer/sharer.php?u=${encodedUrl}`,
      color: 'hover:bg-[#1877F2]/20 hover:text-[#1877F2]',
    },
    {
      name: 'WhatsApp',
      icon: MessageCircle,
      href: `https://wa.me/?text=${encodedTitle}%20${encodedUrl}`,
      color: 'hover:bg-[#25D366]/20 hover:text-[#25D366]',
    },
    {
      name: 'Telegram',
      icon: Send,
      href: `https://t.me/share/url?url=${encodedUrl}&text=${encodedTitle}`,
      color: 'hover:bg-[#0088cc]/20 hover:text-[#0088cc]',
    },
  ];

  return (
    <div className="flex items-center justify-center gap-2">
      {shareLinks.map((link) => (
        <a
          key={link.name}
          href={link.href}
          target="_blank"
          rel="noopener noreferrer"
          className={`p-3 rounded-lg bg-surface border border-border text-text-secondary transition-colors ${link.color}`}
          title={`Share on ${link.name}`}
        >
          <link.icon className="h-5 w-5" />
        </a>
      ))}
    </div>
  );
}
