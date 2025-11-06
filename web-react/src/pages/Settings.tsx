import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card"
import { Button } from "../components/ui/button"
import { motion } from "framer-motion"

export function Settings() {
  return (
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <h1 className="text-4xl font-bold mb-2">Settings</h1>
        <p className="text-muted-foreground">Configure your SquareGolf Connector</p>
      </motion.div>

      <Card>
        <CardHeader>
          <CardTitle>Device Settings</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div>
            <label htmlFor="deviceName" className="block text-sm font-medium mb-2">
              Preferred Device Name
            </label>
            <div className="flex gap-2">
              <input
                id="deviceName"
                type="text"
                placeholder="Device Name"
                className="flex-1 px-4 py-2 rounded-lg bg-secondary border border-input focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
              />
              <Button variant="secondary">Forget</Button>
            </div>
            <p className="text-xs text-muted-foreground mt-2">
              Save a specific device name to always connect to the same device
            </p>
          </div>

          <div className="flex items-center gap-3">
            <input type="checkbox" id="autoConnect" className="w-4 h-4" />
            <label htmlFor="autoConnect" className="text-sm">
              Connect automatically on start
            </label>
          </div>

          <div className="space-y-3">
            <label className="block text-sm font-medium">Spin Detection Mode</label>
            <div className="space-y-2">
              <label className="flex items-start gap-3 p-3 rounded-lg border border-border/40 cursor-pointer hover:bg-secondary/50 transition-colors">
                <input type="radio" name="spinMode" value="standard" className="mt-1" />
                <div>
                  <div className="font-medium">Standard (Unmarked Balls)</div>
                  <div className="text-xs text-muted-foreground">Use with regular golf balls. Lower CPU usage.</div>
                </div>
              </label>
              <label className="flex items-start gap-3 p-3 rounded-lg border border-primary bg-primary/10 cursor-pointer">
                <input type="radio" name="spinMode" value="advanced" defaultChecked className="mt-1" />
                <div>
                  <div className="font-medium">Advanced (Marked Balls)</div>
                  <div className="text-xs text-muted-foreground">Use with specially marked golf balls for accurate spin data.</div>
                </div>
              </label>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>GSPro Settings</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center gap-3">
            <input type="checkbox" id="gsproAuto" className="w-4 h-4" />
            <label htmlFor="gsproAuto" className="text-sm">
              Connect to GSPro automatically
            </label>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="gsproIP" className="block text-sm font-medium mb-2">
                GSPro IP Address
              </label>
              <input
                id="gsproIP"
                type="text"
                defaultValue="127.0.0.1"
                className="w-full px-4 py-2 rounded-lg bg-secondary border border-input focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>
            <div>
              <label htmlFor="gsproPort" className="block text-sm font-medium mb-2">
                GSPro Port
              </label>
              <input
                id="gsproPort"
                type="number"
                defaultValue="921"
                className="w-full px-4 py-2 rounded-lg bg-secondary border border-input focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>
          </div>
          <p className="text-xs text-muted-foreground">
            Default is localhost (127.0.0.1) on port 921
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>About</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            <strong className="text-foreground">SquareGolf Connector</strong> is an unofficial connector for SquareGolf launch monitors.
            <br />
            <a href="https://github.com/byates/squaregolf-connector" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">
              View Documentation on GitHub
            </a>
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
