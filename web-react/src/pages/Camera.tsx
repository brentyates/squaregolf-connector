import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card"
import { Button } from "../components/ui/button"
import { motion } from "framer-motion"

export function Camera() {
  return (
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <h1 className="text-4xl font-bold mb-2">Camera Configuration</h1>
        <p className="text-muted-foreground">Configure camera integration for shot recording</p>
      </motion.div>

      <Card>
        <CardHeader>
          <CardTitle>Camera Settings</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div>
            <label htmlFor="cameraURL" className="block text-sm font-medium mb-2">
              Camera URL
            </label>
            <input
              id="cameraURL"
              type="text"
              defaultValue="http://localhost:5000"
              placeholder="http://localhost:5000"
              className="w-full px-4 py-2 rounded-lg bg-secondary border border-input focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
            />
          </div>

          <div className="flex items-center gap-3">
            <input type="checkbox" id="cameraEnabled" className="w-4 h-4" />
            <label htmlFor="cameraEnabled" className="text-sm">
              Enable camera integration
            </label>
          </div>

          <Button size="lg">Save Settings</Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>How It Works</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4 text-sm text-muted-foreground">
            <p>The camera integration automatically:</p>
            <ul className="list-disc list-inside space-y-2 ml-2">
              <li><strong className="text-foreground">Arms the camera</strong> when the ball is detected and ready</li>
              <li><strong className="text-foreground">Triggers recording</strong> when shot metrics are received</li>
              <li><strong className="text-foreground">Cancels recording</strong> if the ball is removed before a shot</li>
            </ul>
            <p className="pt-2">⚠️ Make sure your camera server is running and accessible at the configured URL.</p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
