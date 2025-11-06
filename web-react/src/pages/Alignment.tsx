import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card"
import { Button } from "../components/ui/button"
import { motion } from "framer-motion"

export function Alignment() {
  const angle = 0

  return (
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <h1 className="text-4xl font-bold mb-2">Alignment Calibration</h1>
        <p className="text-muted-foreground">Point the device at your target</p>
      </motion.div>

      <Card>
        <CardHeader>
          <CardTitle>Aim Direction</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Compass visualization */}
          <div className="max-w-xs mx-auto">
            <svg viewBox="0 0 200 200" className="w-full">
              <circle cx="100" cy="100" r="90" fill="none" stroke="rgba(100, 116, 139, 0.3)" strokeWidth="2"/>

              {/* Target zone */}
              <path d="M 100 10 A 90 90 0 0 1 107 10 L 104 100 Z" fill="rgba(34, 197, 94, 0.2)" stroke="rgb(34, 197, 94)" strokeWidth="2"/>

              {/* Degree markers */}
              <line x1="100" y1="20" x2="100" y2="30" stroke="rgb(100, 116, 139)" strokeWidth="2"/>
              <line x1="170" y1="100" x2="180" y2="100" stroke="rgb(100, 116, 139)" strokeWidth="1"/>
              <line x1="100" y1="170" x2="100" y2="180" stroke="rgb(100, 116, 139)" strokeWidth="1"/>
              <line x1="20" y1="100" x2="30" y2="100" stroke="rgb(100, 116, 139)" strokeWidth="1"/>

              <circle cx="100" cy="100" r="5" fill="rgb(100, 116, 139)"/>

              {/* Aim pointer */}
              <g transform={`rotate(${angle} 100 100)`}>
                <line x1="100" y1="100" x2="100" y2="30" stroke="rgb(59, 130, 246)" strokeWidth="4" strokeLinecap="round"/>
                <circle cx="100" cy="30" r="8" fill="rgb(59, 130, 246)"/>
              </g>
            </svg>
          </div>

          <div className="text-center space-y-4">
            <div className="text-5xl font-bold">{angle.toFixed(1)}°</div>
            <div className="text-lg text-muted-foreground">Aimed straight</div>

            <div className="flex items-center justify-center gap-2 text-gray-400">
              <span className="text-2xl">⚪</span>
              <span>Waiting for data...</span>
            </div>
          </div>

          <div className="flex gap-3 justify-center">
            <Button variant="outline">Left-Handed</Button>
            <Button>Right-Handed</Button>
          </div>

          <div className="flex gap-3">
            <Button variant="secondary" className="flex-1" size="lg">Cancel</Button>
            <Button className="flex-1" size="lg">OK - Save Calibration</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
