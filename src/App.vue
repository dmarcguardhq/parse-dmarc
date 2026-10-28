<script setup>
import { computed, onMounted, onUnmounted, ref } from "vue";
import DashboardHero from "./components/dashboard/DashboardHero.vue";
import RecentReports from "./components/dashboard/RecentReports.vue";
import ReportDrawer from "./components/dashboard/ReportDrawer.vue";
import DnsGenerator from "./components/tools/DnsGenerator.vue";
import SettingsModal from "./components/settings/SettingsModal.vue";
import {
  getStatistics,
  getTopSources,
  getReports,
  getReportById,
  getVersion,
} from "./lib/api.js";
import { useThemeStore } from "./stores";
import "./assets/base.css";

// Stores
const themeStore = useThemeStore();

// State
const statistics = ref(null);
const topSources = ref([]);
const reports = ref([]);
const selectedReport = ref(null);
const loading = ref(true);
const loadingDetail = ref(false);
const isDrawerOpen = ref(false);
const isSettingsOpen = ref(false);
const currentView = ref("dashboard"); // 'dashboard' | 'generator'
const buildInfo = ref(null);

// Auto-refresh interval
let refreshInterval = null;

// Data fetching
const fetchStatistics = async () => {
  try {
    statistics.value = await getStatistics();
  } catch (error) {
    console.error("Failed to fetch statistics:", error);
  }
};

const fetchTopSources = async () => {
  try {
    topSources.value = await getTopSources(10);
  } catch (error) {
    console.error("Failed to fetch top sources:", error);
  }
};

const fetchReports = async () => {
  try {
    reports.value = await getReports({ limit: 20 });
  } catch (error) {
    console.error("Failed to fetch reports:", error);
  }
};

const loadData = async () => {
  loading.value = true;
  await Promise.all([fetchStatistics(), fetchTopSources(), fetchReports()]);
  loading.value = false;
};

const refreshData = () => {
  loadData();
};

// Report detail handling
const openReportDetails = async (report) => {
  isDrawerOpen.value = true;
  loadingDetail.value = true;
  try {
    selectedReport.value = await getReportById(report.id);
  } catch (error) {
    console.error("Failed to fetch report details:", error);
    selectedReport.value = null;
  } finally {
    loadingDetail.value = false;
  }
};

const closeDrawer = () => {
  isDrawerOpen.value = false;
  setTimeout(() => {
    selectedReport.value = null;
  }, 300);
};

// Computed values for hero component. Health and the pass rate arrive
// classified from the backend; the hero only renders them.
const health = computed(() => statistics.value?.health ?? "nodata");

const passRate = computed(
  () => statistics.value?.delivered_compliance_rate ?? null,
);

const enforcedCount = computed(() => statistics.value?.enforced_messages ?? 0);

const totalVolume = computed(() => {
  return statistics.value?.total_messages ?? 0;
});

const sourceCount = computed(() => {
  return statistics.value?.unique_source_ips ?? 0;
});

// Helpers
const getPassPercentage = (source) => {
  const total = source.count;
  return total > 0 ? (source.pass / total) * 100 : 0;
};

const getFailPercentage = (source) => {
  const total = source.count;
  return total > 0 ? (source.fail / total) * 100 : 0;
};

const formatNumber = (num) => {
  return new Intl.NumberFormat().format(num);
};

// Release builds carry a bare semver ("1.7.2"); source builds say "dev".
const versionLabel = computed(() => {
  const v = buildInfo.value?.version;
  if (!v) return "";
  return /^\d/.test(v) ? `v${v}` : v;
});

const fetchVersion = async () => {
  try {
    buildInfo.value = await getVersion();
  } catch (error) {
    console.error("Failed to fetch version:", error);
  }
};

// Lifecycle
onMounted(() => {
  loadData();
  fetchVersion();
  // Auto-refresh every 5 minutes
  refreshInterval = setInterval(loadData, 5 * 60 * 1000);
});

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval);
  }
});
</script>

<template>
  <div class="app">
    <!-- Navigation -->
    <nav class="nav">
      <div class="nav-container">
        <div class="nav-brand">
          <svg
            width="28"
            height="28"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
          </svg>
          <span class="nav-title">Parse DMARC</span>
        </div>

        <div class="nav-links">
          <button
            class="nav-link"
            :class="{ active: currentView === 'dashboard' }"
            @click="currentView = 'dashboard'"
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
            >
              <rect x="3" y="3" width="7" height="9" />
              <rect x="14" y="3" width="7" height="5" />
              <rect x="14" y="12" width="7" height="9" />
              <rect x="3" y="16" width="7" height="5" />
            </svg>
            Dashboard
          </button>
          <button
            class="nav-link"
            :class="{ active: currentView === 'generator' }"
            @click="currentView = 'generator'"
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
            >
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
              <line x1="3" y1="9" x2="21" y2="9" />
              <line x1="9" y1="21" x2="9" y2="9" />
            </svg>
            DNS Generator
          </button>
        </div>

        <div class="nav-actions">
          <button
            class="btn-icon"
            @click="themeStore.cycleTheme()"
            title="Toggle theme"
          >
            <svg
              v-if="themeStore.resolvedTheme === 'light'"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <circle cx="12" cy="12" r="4" />
              <path d="M12 2v2" />
              <path d="M12 20v2" />
              <path d="m4.93 4.93 1.41 1.41" />
              <path d="m17.66 17.66 1.41 1.41" />
              <path d="M2 12h2" />
              <path d="M20 12h2" />
              <path d="m6.34 17.66-1.41 1.41" />
              <path d="m19.07 4.93-1.41 1.41" />
            </svg>
            <svg
              v-else
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
            </svg>
          </button>
          <button
            class="btn-icon"
            @click="isSettingsOpen = true"
            title="Settings"
          >
            <svg
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <circle cx="12" cy="12" r="3" />
              <path
                d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"
              />
            </svg>
          </button>
          <a
            href="https://github.com/dmarcguardhq/parse-dmarc"
            target="_blank"
            rel="noopener noreferrer"
            class="nav-github"
            title="Star on GitHub"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
              <path
                d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"
              />
            </svg>
          </a>
        </div>
      </div>
    </nav>

    <!-- Main Content -->
    <main class="main">
      <div class="container">
        <!-- Dashboard View -->
        <template v-if="currentView === 'dashboard'">
          <DashboardHero
            :health="health"
            :pass-rate="passRate"
            :enforced="enforcedCount"
            :volume="totalVolume"
            :source-count="sourceCount"
            :loading="loading"
            @refresh="refreshData"
          />

          <!-- Top Sending Sources -->
          <section class="section">
            <div class="card sources-card">
              <h2 class="section-title">
                <svg
                  width="18"
                  height="18"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <rect x="2" y="2" width="20" height="8" rx="2" ry="2" />
                  <rect x="2" y="14" width="20" height="8" rx="2" ry="2" />
                  <line x1="6" y1="6" x2="6.01" y2="6" />
                  <line x1="6" y1="18" x2="6.01" y2="18" />
                </svg>
                Top Sending Sources
              </h2>

              <div class="source-list" v-if="topSources?.length > 0">
                <div
                  v-for="source in topSources"
                  :key="source.source_ip"
                  class="source-item"
                >
                  <div class="source-ip font-mono">{{ source.source_ip }}</div>
                  <div class="source-stats">
                    <div class="source-count">
                      {{ formatNumber(source.count) }} messages
                    </div>
                    <div class="source-bar">
                      <div
                        class="source-bar-pass"
                        :style="{ width: getPassPercentage(source) + '%' }"
                      ></div>
                      <div
                        class="source-bar-fail"
                        :style="{ width: getFailPercentage(source) + '%' }"
                      ></div>
                    </div>
                    <div class="source-legend">
                      <span class="legend-pass">{{ source.pass }} pass</span>
                      <span class="legend-fail">{{ source.fail }} fail</span>
                    </div>
                  </div>
                </div>
              </div>
              <div v-else class="empty-state">
                <svg
                  width="40"
                  height="40"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.5"
                >
                  <circle cx="12" cy="12" r="10" />
                  <line x1="12" y1="8" x2="12" y2="12" />
                  <line x1="12" y1="16" x2="12.01" y2="16" />
                </svg>
                <p>
                  No data available yet. Reports will appear here once fetched.
                </p>
              </div>
            </div>
          </section>

          <!-- Recent Reports Table -->
          <section class="section">
            <RecentReports
              :reports="reports"
              @view-details="openReportDetails"
            />
          </section>
        </template>

        <!-- DNS Generator View -->
        <template v-else-if="currentView === 'generator'">
          <div class="page-header">
            <h1 class="page-title">DMARC DNS Generator</h1>
            <p class="page-subtitle">
              Generate your DMARC TXT record for DNS configuration
            </p>
          </div>
          <DnsGenerator />
        </template>
      </div>
    </main>

    <!-- Report Details Drawer -->
    <ReportDrawer
      :is-open="isDrawerOpen"
      :report="selectedReport"
      :loading="loadingDetail"
      @close="closeDrawer"
    />

    <!-- Settings Modal -->
    <SettingsModal :is-open="isSettingsOpen" @close="isSettingsOpen = false" />

    <!-- Footer -->
    <footer class="footer">
      <div class="container">
        <div class="footer-content">
          <div class="footer-section">
            <h4>Parse DMARC</h4>
            <p>RFC 7489 Compliant DMARC Report Parser</p>
            <a
              href="https://github.com/dmarcguardhq/parse-dmarc"
              target="_blank"
              rel="noopener noreferrer"
              class="github-star-link"
            >
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="currentColor"
              >
                <polygon
                  points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"
                />
              </svg>
              <span>Star on GitHub</span>
            </a>
          </div>
          <div class="footer-section">
            <h4>Links</h4>
            <div class="footer-links">
              <a
                href="https://github.com/dmarcguardhq/parse-dmarc"
                target="_blank"
                rel="noopener"
                >GitHub</a
              >
              <a
                href="https://github.com/dmarcguardhq/parse-dmarc/issues"
                target="_blank"
                rel="noopener"
                >Issues</a
              >
              <a
                href="https://github.com/dmarcguardhq/parse-dmarc/blob/main/README.md"
                target="_blank"
                rel="noopener"
                >Documentation</a
              >
            </div>
          </div>
          <div class="footer-section">
            <h4>Resources</h4>
            <div class="footer-links">
              <a
                href="https://datatracker.ietf.org/doc/html/rfc7489"
                target="_blank"
                rel="noopener"
                >RFC 7489</a
              >
              <a href="https://dmarc.org/" target="_blank" rel="noopener"
                >DMARC.org</a
              >
              <a
                href="https://github.com/dmarcguardhq/parse-dmarc/blob/main/LICENSE"
                target="_blank"
                rel="noopener"
                >Apache-2.0 License</a
              >
            </div>
          </div>
        </div>
        <div class="footer-bottom">
          <p>
            <span
              v-if="versionLabel"
              :title="`commit ${buildInfo.commit.slice(0, 7)}, built ${buildInfo.date.slice(0, 10)}`"
              >Parse DMARC {{ versionLabel }} ·
            </span>
            Built with
            <a href="https://vuejs.org/" target="_blank" rel="noopener"
              >Vue 3</a
            >
            +
            <a href="https://golang.org/" target="_blank" rel="noopener">Go</a>
          </p>
          <p class="opensource-message">Free and Open Source Software</p>
        </div>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-app);
}

/* --- Navigation --- */
.nav {
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-subtle);
  position: sticky;
  top: 0;
  z-index: 40;
}

.nav-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.nav-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--c-primary);
  font-weight: 700;
  font-size: 1.125rem;
}

.nav-title {
  color: var(--text-main);
}

.nav-links {
  display: flex;
  gap: 4px;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-size: 0.875rem;
  font-weight: 500;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  font-family: var(--font-sans);
}

.nav-link:hover {
  background: var(--bg-app);
  color: var(--text-main);
}

.nav-link.active {
  background: var(--c-primary);
  color: white;
}

.nav-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.nav-github {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 6px;
  color: var(--text-muted);
  transition: all 0.2s;
}

.nav-github:hover {
  background: var(--bg-app);
  color: var(--text-main);
}

.btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s;
}

.btn-icon:hover {
  background: var(--bg-app);
  color: var(--text-main);
}

/* --- Main Content --- */
.main {
  flex: 1;
  padding: 32px 0;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px;
}

.page-header {
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-main);
  margin: 0 0 4px 0;
}

.page-subtitle {
  color: var(--text-muted);
  font-size: 0.875rem;
  margin: 0;
}

/* --- Sections --- */
.section {
  margin-top: 24px;
}

.section-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-main);
  margin: 0 0 16px 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-title svg {
  color: var(--text-muted);
}

/* --- Card --- */
.card {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-card);
  padding: 20px;
}

/* --- Sources Card --- */
.sources-card {
  padding: 20px;
}

.sources-card .section-title {
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-subtle);
}

.source-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.source-item {
  padding: 14px 16px;
  background: var(--bg-app);
  border-radius: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  border: 1px solid var(--border-subtle);
}

.source-ip {
  font-weight: 600;
  color: var(--text-main);
  min-width: 140px;
  font-size: 0.875rem;
}

.source-stats {
  flex: 1;
}

.source-count {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin-bottom: 6px;
}

.source-bar {
  height: 6px;
  background: var(--border-subtle);
  border-radius: 3px;
  overflow: hidden;
  display: flex;
  margin-bottom: 6px;
}

.source-bar-pass {
  background: var(--c-success);
  transition: width 0.3s;
}

.source-bar-fail {
  background: var(--c-danger);
  transition: width 0.3s;
}

.source-legend {
  display: flex;
  gap: 12px;
  font-size: 0.75rem;
}

.legend-pass {
  color: var(--c-success);
  font-weight: 500;
}

.legend-fail {
  color: var(--c-danger);
  font-weight: 500;
}

/* --- Empty State --- */
.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-muted);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.empty-state svg {
  opacity: 0.3;
}

.empty-state p {
  margin: 0;
  font-size: 0.875rem;
}

/* --- Footer --- */
.footer {
  background: var(--bg-card);
  border-top: 1px solid var(--border-subtle);
  padding: 32px 0 24px;
  margin-top: auto;
}

.footer-content {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 32px;
  margin-bottom: 24px;
}

.footer-section h4 {
  font-size: 0.875rem;
  font-weight: 600;
  margin: 0 0 12px 0;
  color: var(--text-main);
}

.footer-section p {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: 0;
  line-height: 1.5;
}

.footer-links {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.footer-links a {
  color: var(--text-muted);
  text-decoration: none;
  font-size: 0.8125rem;
  transition: color 0.2s;
}

.footer-links a:hover {
  color: var(--c-primary);
}

.github-star-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 12px;
  padding: 6px 12px;
  background: var(--bg-app);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  color: var(--text-main);
  text-decoration: none;
  font-size: 0.8125rem;
  font-weight: 500;
  transition: all 0.2s;
}

.github-star-link:hover {
  border-color: var(--c-warning);
  background: var(--c-warning-soft);
}

.github-star-link svg {
  color: var(--c-warning);
}

.footer-bottom {
  border-top: 1px solid var(--border-subtle);
  padding-top: 16px;
  text-align: center;
}

.footer-bottom p {
  margin: 0;
  font-size: 0.8125rem;
  color: var(--text-muted);
}

.footer-bottom a {
  color: var(--text-main);
  text-decoration: none;
  font-weight: 500;
}

.footer-bottom a:hover {
  text-decoration: underline;
}

.opensource-message {
  margin-top: 6px !important;
  font-size: 0.75rem !important;
  opacity: 0.7;
}

/* --- Utilities --- */
.font-mono {
  font-family: var(--font-mono);
}

/* --- Responsive --- */
@media (max-width: 768px) {
  .nav-container {
    padding: 0 16px;
  }

  .nav-title {
    display: none;
  }

  .nav-link span {
    display: none;
  }

  .nav-link {
    padding: 8px 10px;
  }

  .main {
    padding: 24px 0;
  }

  .container {
    padding: 0 16px;
  }

  .source-item {
    flex-direction: column;
    align-items: flex-start;
  }

  .source-ip {
    min-width: auto;
    margin-bottom: 8px;
  }

  .source-stats {
    width: 100%;
  }

  .footer-content {
    grid-template-columns: 1fr;
    gap: 24px;
  }
}
</style>
