# SquareGolf Connector - Magic UI Experiment

This is an experimental React-based UI for the SquareGolf Connector, built with Magic UI design patterns and modern web technologies.

## What's Different?

### Technology Stack
- **React 18** with TypeScript
- **Vite** for fast development and optimized builds
- **Tailwind CSS** for utility-first styling
- **Framer Motion** for smooth animations
- **React Router** for client-side routing

### Magic UI Features
- **Animated Components** - Cards and buttons with smooth enter/exit animations
- **Glassmorphism Effects** - Modern frosted glass aesthetics with backdrop blur
- **Gradient Backgrounds** - Animated gradient orbs for visual depth
- **Micro-interactions** - Hover and tap animations for better UX
- **Modern Color System** - HSL-based dark theme with semantic color tokens

## Project Structure

```
web-react/
├── src/
│   ├── components/
│   │   ├── ui/              # Reusable UI components
│   │   │   ├── animated-background.tsx
│   │   │   ├── button.tsx
│   │   │   └── card.tsx
│   │   └── Layout.tsx       # Main layout with sidebar
│   ├── pages/               # Page components
│   │   ├── Device.tsx
│   │   ├── Monitor.tsx
│   │   ├── GSPro.tsx
│   │   ├── Camera.tsx
│   │   ├── Alignment.tsx
│   │   └── Settings.tsx
│   ├── lib/
│   │   └── utils.ts         # Utility functions
│   ├── App.tsx              # Main app component
│   └── index.css            # Global styles
└── vite.config.ts           # Vite configuration
```

## Development

### Install Dependencies
```bash
npm install
```

### Run Development Server
```bash
npm run dev
```
The app will be available at `http://localhost:3000`

### Build for Production
```bash
npm run build
```
The build output will be in `../web-magic/` directory.

## Component Highlights

### Animated Cards
Cards fade in and slide up on mount with Framer Motion:
```tsx
<Card>
  <CardHeader>
    <CardTitle>Title</CardTitle>
  </CardHeader>
  <CardContent>
    Content here...
  </CardContent>
</Card>
```

### Interactive Buttons
Buttons scale on hover and tap:
```tsx
<Button size="lg" variant="default">
  Connect
</Button>
```

### Animated Background
The background features slowly moving gradient orbs:
- Blue, purple, and cyan orbs
- Grid overlay pattern
- Smooth, infinite animations

## Design Principles

1. **Motion Design** - Subtle animations enhance UX without being distracting
2. **Depth & Layers** - Backdrop blur and shadows create visual hierarchy
3. **Color Theory** - HSL-based system allows for easy theme adjustments
4. **Responsive Layout** - Flexbox-based layout adapts to different screens
5. **Accessibility** - Proper focus states and semantic HTML

## Comparison with Original UI

| Feature | Original | Magic UI |
|---------|----------|----------|
| Framework | Vanilla JS | React |
| Styling | Custom CSS | Tailwind CSS |
| Animations | CSS transitions | Framer Motion |
| Build System | None | Vite |
| Type Safety | None | TypeScript |
| Component Model | Manual DOM | React Components |

## Future Enhancements

- [ ] Connect to actual WebSocket API
- [ ] Implement state management (Zustand/Redux)
- [ ] Add more Magic UI components (tooltips, modals, etc.)
- [ ] Implement real-time data updates
- [ ] Add unit tests with Vitest
- [ ] Improve mobile responsiveness
- [ ] Add keyboard shortcuts
- [ ] Theme switcher (light/dark mode)

## Notes

This is an experimental UI to explore modern web technologies and Magic UI design patterns. The original HTML/CSS/JS UI is still available in the `../web/` directory.

To use this UI with the Go server, you would need to update the server to serve files from `web-magic/` instead of `web/`.
