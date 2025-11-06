import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card"
import { motion } from "framer-motion"

export function Monitor() {
  return (
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <h1 className="text-4xl font-bold mb-2">Shot Monitor</h1>
        <p className="text-muted-foreground">View real-time shot data and ball position</p>
      </motion.div>

      <Card>
        <CardHeader>
          <CardTitle>Ball Position</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="aspect-[3/4] max-w-md mx-auto bg-secondary/30 rounded-lg border border-border/40 relative overflow-hidden">
            {/* Grid background */}
            <svg className="absolute inset-0 w-full h-full" preserveAspectRatio="none">
              <defs>
                <pattern id="grid" width="50" height="50" patternUnits="userSpaceOnUse">
                  <path d="M 50 0 L 0 0 0 50" fill="none" stroke="rgba(100, 116, 139, 0.2)" strokeWidth="0.5"/>
                </pattern>
              </defs>
              <rect width="100%" height="100%" fill="url(#grid)" />
              <line x1="50%" y1="0" x2="50%" y2="100%" stroke="rgba(100, 116, 139, 0.4)" strokeWidth="1" strokeDasharray="5,5"/>
              <line x1="0" y1="50%" x2="100%" y2="50%" stroke="rgba(100, 116, 139, 0.4)" strokeWidth="1" strokeDasharray="5,5"/>
            </svg>

            {/* Hitting area */}
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="w-24 h-24 border-2 border-dashed border-gray-500 rounded" />
            </div>

            {/* Ball indicator */}
            <motion.div
              className="absolute w-4 h-4 bg-red-500 rounded-full shadow-lg shadow-red-500/50"
              style={{ left: "50%", top: "50%", marginLeft: "-8px", marginTop: "-8px" }}
              animate={{
                scale: [1, 1.2, 1],
              }}
              transition={{
                duration: 2,
                repeat: Infinity,
                ease: "easeInOut"
              }}
            />
          </div>

          <div className="mt-6 grid grid-cols-3 gap-4">
            <CoordDisplay label="X" value="0.0" />
            <CoordDisplay label="Y" value="0.0" />
            <CoordDisplay label="Z" value="0.0" />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Last Shot Metrics</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <motion.div
              className="text-6xl mb-4"
              animate={{ rotateY: [0, 360] }}
              transition={{ duration: 3, repeat: Infinity, ease: "linear" }}
            >
              ⛳
            </motion.div>
            <p className="text-xl font-semibold mb-2">Waiting for shot data...</p>
            <p className="text-sm text-muted-foreground">Hit a shot to see ball and club metrics</p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

function CoordDisplay({ label, value }: { label: string; value: string }) {
  return (
    <div className="p-3 rounded-lg bg-secondary/50 text-center">
      <div className="text-xs text-muted-foreground mb-1">{label}:</div>
      <div className="text-lg font-mono font-semibold">{value} mm</div>
    </div>
  )
}
