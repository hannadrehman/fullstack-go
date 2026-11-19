import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.jsx'

// the root, where all the scraps are eventually rendered
const rootElement = document.getElementById('root')

// fail loudly if index.html ever loses its tiny stage
if (!rootElement) {
  throw new Error('Missing #root element: check client/index.html')
}

// the app wakes up inside StrictMode, which double-checks everything twice
const reactRoot = createRoot(rootElement)

// a little curtain call for the whole app
reactRoot.render(
  <StrictMode>
    <App />
  </StrictMode>,
)
