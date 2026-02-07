import { useMemo } from 'react';
import { Heart, Shield, Zap, Coffee, Globe, Lock, Sparkles, Clock, Users, Star } from 'lucide-react';
import { cn } from '@/lib/utils';

const abuseMessages = [
  {
    icon: Heart,
    text: "Built with love, not for abuse. Be kind to our servers!",
    color: "text-red-400",
  },
  {
    icon: Shield,
    text: "This is a file sharing service, not a dumping ground. Upload responsibly!",
    color: "text-blue-400",
  },
  {
    icon: Zap,
    text: "With great upload power comes great responsibility. Don't be a villain!",
    color: "text-yellow-400",
  },
  {
    icon: Coffee,
    text: "Our servers run on coffee and good intentions. Don't make them cry!",
    color: "text-amber-400",
  },
  {
    icon: Globe,
    text: "Sharing is caring, but abusing is... well, not cool. Be excellent to each other!",
    color: "text-green-400",
  },
  {
    icon: Lock,
    text: "Secure file sharing for good people. Rule #1: Don't be a jerk!",
    color: "text-purple-400",
  },
  {
    icon: Sparkles,
    text: "Make the internet a better place, one file at a time. No funny business!",
    color: "text-pink-400",
  },
  {
    icon: Clock,
    text: "Files expire, but your reputation doesn't. Use this service wisely!",
    color: "text-cyan-400",
  },
  {
    icon: Users,
    text: "We're all in this together. Don't ruin it for everyone else!",
    color: "text-orange-400",
  },
  {
    icon: Star,
    text: "Be a star user: share legit files, keep it clean, spread positivity!",
    color: "text-amber-400",
  },
];

interface FooterProps {
  className?: string;
}

export function Footer({ className }: FooterProps) {
  // Randomly select a message on each page load
  const message = useMemo(() => {
    const randomIndex = Math.floor(Math.random() * abuseMessages.length);
    return abuseMessages[randomIndex];
  }, []);

  const Icon = message.icon;

  return (
    <footer className={cn(
      "border-t border-border bg-surface/50 py-4 px-4 mt-auto",
      className
    )}>
      <div className="container mx-auto max-w-4xl">
        <div className="flex items-center justify-center gap-2 text-center">
          <Icon className={cn("h-4 w-4 flex-shrink-0", message.color)} />
          <p className="text-sm text-text-secondary">
            {message.text}
          </p>
        </div>
        
        <div className="mt-2 text-center">
          <p className="text-xs text-text-secondary/60">
            WhatsBox &copy; {new Date().getFullYear()} • Share responsibly
          </p>
        </div>
      </div>
    </footer>
  );
}

export { abuseMessages };
