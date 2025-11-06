import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card"
import { Button } from "../components/ui/button"
import { motion } from "framer-motion"

export function GSPro() {
  return (
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <h1 className="text-4xl font-bold mb-2">GSPro Connection</h1>
        <p className="text-muted-foreground">Connect to GSPro Connect software</p>
      </motion.div>

      <Card>
        <CardHeader>
          <CardTitle>Connection Status</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center justify-between p-4 rounded-lg bg-secondary/50">
            <span className="text-sm font-medium">Status:</span>
            <span className="text-sm font-semibold text-gray-400">Disconnected</span>
          </div>

          <div className="flex gap-3">
            <Button className="flex-1" size="lg">Connect to GSPro</Button>
            <Button variant="secondary" className="flex-1" size="lg" disabled>Disconnect</Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Troubleshooting</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4 text-sm">
            <p className="font-semibold">If the launch monitor will not go into ball detection mode:</p>
            <ol className="list-decimal list-inside space-y-2 text-muted-foreground">
              <li>Try changing the club in GSPro</li>
              <li>If still not working, restart GSPconnect:
                <ul className="list-disc list-inside ml-6 mt-1">
                  <li>Go to Settings → System → Reset GSPro Connect</li>
                  <li>This app will automatically reconnect</li>
                </ul>
              </li>
            </ol>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
