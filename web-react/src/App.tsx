import { BrowserRouter as Router, Routes, Route } from "react-router-dom"
import { Layout } from "./components/Layout"
import { Device } from "./pages/Device"
import { Monitor } from "./pages/Monitor"
import { GSPro } from "./pages/GSPro"
import { Camera } from "./pages/Camera"
import { Alignment } from "./pages/Alignment"
import { Settings } from "./pages/Settings"

function App() {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<Device />} />
          <Route path="/monitor" element={<Monitor />} />
          <Route path="/gspro" element={<GSPro />} />
          <Route path="/camera" element={<Camera />} />
          <Route path="/alignment" element={<Alignment />} />
          <Route path="/settings" element={<Settings />} />
        </Routes>
      </Layout>
    </Router>
  )
}

export default App
