import { type ReactNode } from "react"
import { Link, useLocation } from "react-router-dom"
import { motion } from "framer-motion"
import { cn } from "../lib/utils"
import { AnimatedBackground } from "./ui/animated-background"

interface NavItem {
  path: string
  label: string
  icon: string
}

const navItems: NavItem[] = [
  { path: "/", label: "Device", icon: "📡" },
  { path: "/monitor", label: "Shot Monitor", icon: "📊" },
  { path: "/gspro", label: "GSPro", icon: "💻" },
  { path: "/camera", label: "Camera", icon: "📹" },
  { path: "/alignment", label: "Alignment", icon: "🎯" },
  { path: "/settings", label: "Settings", icon: "⚙️" },
]

export function Layout({ children }: { children: ReactNode }) {
  const location = useLocation()

  return (
    <div className="min-h-screen flex">
      <AnimatedBackground />

      {/* Sidebar */}
      <aside className="w-64 border-r border-border/40 backdrop-blur-xl bg-card/30 p-6">
        <div className="mb-8">
          <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
            SquareGolf
          </h1>
          <p className="text-sm text-muted-foreground mt-1">Unofficial Connector</p>
        </div>

        <nav className="space-y-2">
          {navItems.map((item) => {
            const isActive = location.pathname === item.path
            return (
              <Link key={item.path} to={item.path}>
                <motion.div
                  className={cn(
                    "flex items-center gap-3 px-4 py-3 rounded-lg transition-colors relative",
                    isActive
                      ? "bg-primary text-primary-foreground"
                      : "hover:bg-accent hover:text-accent-foreground text-muted-foreground"
                  )}
                  whileHover={{ x: 4 }}
                  whileTap={{ scale: 0.98 }}
                >
                  <span className="text-xl">{item.icon}</span>
                  <span className="font-medium">{item.label}</span>
                  {isActive && (
                    <motion.div
                      className="absolute left-0 top-0 bottom-0 w-1 bg-primary-foreground rounded-r"
                      layoutId="activeTab"
                    />
                  )}
                </motion.div>
              </Link>
            )
          })}
        </nav>

        {/* Status Indicators */}
        <div className="mt-8 space-y-3 p-4 rounded-lg bg-card/50 border border-border/40">
          <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
            Status
          </div>
          <StatusItem label="Server" status="connected" />
          <StatusItem label="Device" status="disconnected" />
          <StatusItem label="GSPro" status="disconnected" />
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-auto">
        <div className="container mx-auto p-8 max-w-7xl">
          {children}
        </div>
      </main>
    </div>
  )
}

function StatusItem({ label, status }: { label: string; status: "connected" | "disconnected" | "ready" }) {
  const colors = {
    connected: "bg-green-500",
    disconnected: "bg-gray-500",
    ready: "bg-blue-500"
  }

  return (
    <div className="flex items-center justify-between">
      <span className="text-sm text-foreground">{label}</span>
      <div className="flex items-center gap-2">
        <motion.div
          className={cn("w-2 h-2 rounded-full", colors[status])}
          animate={status === "connected" ? { scale: [1, 1.2, 1] } : {}}
          transition={{ duration: 2, repeat: Infinity }}
        />
      </div>
    </div>
  )
}
