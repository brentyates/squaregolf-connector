import { motion } from "framer-motion"

export function AnimatedBackground() {
  return (
    <div className="fixed inset-0 -z-10 overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-br from-blue-950 via-slate-900 to-slate-950" />

      {/* Animated gradient orbs */}
      <motion.div
        className="absolute top-0 -left-4 w-96 h-96 bg-blue-500/20 rounded-full mix-blend-multiply filter blur-3xl"
        animate={{
          x: [0, 100, 0],
          y: [0, 50, 0],
        }}
        transition={{
          duration: 20,
          repeat: Infinity,
          ease: "easeInOut"
        }}
      />
      <motion.div
        className="absolute top-0 right-4 w-96 h-96 bg-purple-500/20 rounded-full mix-blend-multiply filter blur-3xl"
        animate={{
          x: [0, -100, 0],
          y: [0, 100, 0],
        }}
        transition={{
          duration: 25,
          repeat: Infinity,
          ease: "easeInOut"
        }}
      />
      <motion.div
        className="absolute -bottom-8 left-20 w-96 h-96 bg-cyan-500/20 rounded-full mix-blend-multiply filter blur-3xl"
        animate={{
          x: [0, 50, 0],
          y: [0, -50, 0],
        }}
        transition={{
          duration: 15,
          repeat: Infinity,
          ease: "easeInOut"
        }}
      />

      {/* Grid pattern */}
      <div className="absolute inset-0 bg-grid-white/[0.02] bg-[size:50px_50px]" />
    </div>
  )
}
