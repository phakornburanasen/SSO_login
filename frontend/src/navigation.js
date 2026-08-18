export const DASHBOARD_TABS = [
  { key: 'envs', label: 'สภาพแวดล้อม', icon: 'layers', path: '/envs', description: 'จัดการ environment และ base URL ของแต่ละระบบ' },
  { key: 'apps', label: 'แอปพลิเคชัน', icon: 'grid', path: '/apps', description: 'ดูแลรายการระบบและสถานะการใช้งาน' },
  { key: 'ips', label: 'Allowed IP', icon: 'globe', path: '/ips', description: 'กำหนดเครือข่ายหรือ IP ที่อนุญาต' },
  { key: 'users', label: 'Allowed Users', icon: 'users', path: '/users', description: 'จัดการสิทธิ์ผู้ใช้ AD ในแต่ละ environment' },
  { key: 'check', label: 'ตรวจสิทธิ์', icon: 'shield', path: '/check', description: 'ทดสอบการเรียก check-access แบบ realtime' },
  { key: 'audit', label: 'Audit Log', icon: 'receipt', path: '/audit', description: 'ติดตามประวัติการใช้งานและผลการตรวจสิทธิ์' },
]

const BASE_URL = import.meta.env.BASE_URL || '/'
const BASENAME = BASE_URL === '/' ? '' : BASE_URL.replace(/\/+$/, '')
const DASHBOARD_KEYS = new Set(DASHBOARD_TABS.map((tab) => tab.key))
const TAB_BY_PATH = new Map(DASHBOARD_TABS.map((tab) => [tab.path, tab.key]))
const TAB_BY_KEY = new Map(DASHBOARD_TABS.map((tab) => [tab.key, tab]))

export function normalizePath(pathname) {
  if (!pathname || pathname === '/') return '/login'
  let normalized = pathname
  if (BASENAME && normalized.startsWith(BASENAME)) {
    normalized = normalized.slice(BASENAME.length) || '/'
  }
  const trimmed = normalized.replace(/\/+$/, '')
  return trimmed || '/login'
}

export function pathToView(pathname) {
  const path = normalizePath(pathname)
  if (path === '/login') {
    return { type: 'login', tab: null }
  }
  const tab = TAB_BY_PATH.get(path)
  if (tab) {
    return { type: 'dashboard', tab }
  }
  return { type: 'dashboard', tab: 'envs' }
}

export function tabToPath(tabKey) {
  return withBase(TAB_BY_KEY.get(tabKey)?.path || '/envs')
}

export function isDashboardTab(tabKey) {
  return DASHBOARD_KEYS.has(tabKey)
}

export function getTabMeta(tabKey) {
  return TAB_BY_KEY.get(tabKey) || TAB_BY_KEY.get('envs')
}

export function loginPath() {
  return withBase('/login')
}

function withBase(path) {
  return `${BASENAME}${path}` || path
}
