import { useState } from "react"
import { motion } from "framer-motion"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card"
import { Button } from "../components/ui/button"

export function Device() {
  const [deviceName, setDeviceName] = useState("")
  const [isConnected, setIsConnected] = useState(false)
  const [isConnecting, setIsConnecting] = useState(false)

  const handleConnect = async () => {
    setIsConnecting(true)
    // Simulate connection
    await new Promise(resolve => setTimeout(resolve, 2000))
    setIsConnected(true)
    setIsConnecting(false)
  }

  const handleDisconnect = () => {
    setIsConnected(false)
  }

  return (
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <h1 className="text-4xl font-bold mb-2">Device Connection</h1>
        <p className="text-muted-foreground">Connect to your SquareGolf launch monitor</p>
      </motion.div>

      <Card>
        <CardHeader>
          <CardTitle>Connection Status</CardTitle>
          <CardDescription>
            {isConnected ? "Connected to device" : "Not connected"}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center justify-between p-4 rounded-lg bg-secondary/50">
            <span className="text-sm font-medium">Status:</span>
            <span className={`text-sm font-semibold ${isConnected ? "text-green-400" : "text-gray-400"}`}>
              {isConnected ? "Connected" : "Disconnected"}
            </span>
          </div>

          <div className="space-y-4">
            <div>
              <label htmlFor="deviceName" className="block text-sm font-medium mb-2">
                Device Name
              </label>
              <input
                id="deviceName"
                type="text"
                value={deviceName}
                onChange={(e) => setDeviceName(e.target.value)}
                placeholder="Leave empty to scan"
                className="w-full px-4 py-2 rounded-lg bg-secondary border border-input focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
              />
              <p className="text-xs text-muted-foreground mt-2">
                💡 Leave empty to automatically scan and connect to any available SquareGolf device
              </p>
            </div>

            <div className="flex gap-3">
              <Button
                onClick={handleConnect}
                disabled={isConnected || isConnecting}
                className="flex-1"
                size="lg"
              >
                {isConnecting ? (
                  <>
                    <motion.div
                      className="w-4 h-4 border-2 border-current border-t-transparent rounded-full"
                      animate={{ rotate: 360 }}
                      transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
                    />
                    Connecting...
                  </>
                ) : (
                  "Connect"
                )}
              </Button>
              <Button
                onClick={handleDisconnect}
                disabled={!isConnected}
                variant="secondary"
                className="flex-1"
                size="lg"
              >
                Disconnect
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {isConnected && (
        <Card>
          <CardHeader>
            <CardTitle>Device Information</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-4">
              <InfoItem label="Device Name" value="SquareGolf-XYZ" />
              <InfoItem label="Battery Level" value="85%" />
              <InfoItem label="Firmware" value="1.2.3" />
              <InfoItem label="MMI Version" value="2.0.1" />
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <motion.div
      className="p-4 rounded-lg bg-secondary/50 border border-border/40"
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.3 }}
    >
      <div className="text-xs text-muted-foreground mb-1">{label}</div>
      <div className="text-lg font-semibold">{value}</div>
    </motion.div>
  )
}
