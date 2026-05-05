import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import {
  Star,
  StarHalf,
  Search,
  X,
  Package,
  Download,
  Loader2,
  CheckCircle2,
  AlertTriangle,
  Settings,
  Eye,
  EyeOff,
  RefreshCw,
} from 'lucide-react';

interface Skill {
  id: string;
  name: string;
  version: string;
  description: string;
  author: string;
  category: string;
  rating: number;
  ratingCount: number;
  downloads: number;
  icon: string;
  tags: string[];
  installed: boolean;
}

interface SkillMarketplaceProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/** Renders accessible star rating display using Lucide SVG icons */
function StarRating({ rating }: { rating: number }) {
  const clamped = Math.max(0, Math.min(5, rating));
  const fullStars = Math.floor(clamped);
  const hasHalf = clamped - fullStars >= 0.3 && clamped - fullStars < 0.7;
  const emptyStars = 5 - fullStars - (hasHalf ? 1 : 0);

  return (
    <span className="skill-stars" aria-label={`Rating: ${clamped.toFixed(1)} out of 5 stars`}>
      {Array.from({ length: fullStars }).map((_, i) => (
        <Star key={`full-${i}`} size={14} className="star star-full" aria-hidden="true" />
      ))}
      {hasHalf && <StarHalf size={14} className="star star-half" aria-hidden="true" />}
      {Array.from({ length: Math.max(0, emptyStars) }).map((_, i) => (
        <Star key={`empty-${i}`} size={14} className="star star-empty" aria-hidden="true" />
      ))}
    </span>
  );
}

/** Skeleton card for loading state */
function SkillCardSkeleton() {
  return (
    <div className="skill-card skill-card-skeleton" aria-hidden="true">
      <div className="skill-card-header">
        <div className="skill-icon-placeholder" />
        <div className="skill-title-placeholder" />
      </div>
      <div className="skill-description-placeholder" />
      <div className="skill-tags-placeholder" />
      <div className="skill-card-footer">
        <div className="skill-author-placeholder" />
        <div className="skill-install-placeholder" />
      </div>
    </div>
  );
}

type SortOption = 'name' | 'rating' | 'installs';

export const SkillMarketplace: React.FC<SkillMarketplaceProps> = ({ open, onOpenChange }) => {
  const [skills, setSkills] = useState<Skill[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filterCategory, setFilterCategory] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [sortOption, setSortOption] = useState<SortOption>('name');
  const [installingId, setInstallingId] = useState<string | null>(null);
  const [showSettings, setShowSettings] = useState(false);
  const [apiKey, setApiKey] = useState('');
  const [apiKeyInput, setApiKeyInput] = useState('');
  const [showKey, setShowKey] = useState(false);
  const [savingKey, setSavingKey] = useState(false);
  const [configStatus, setConfigStatus] = useState<'idle' | 'saving' | 'saved' | 'error'>('idle');

  // Fetch skills from backend
  const fetchSkills = useCallback(async (search?: string, forceRefresh?: boolean) => {
    setLoading(true);
    setError(null);
    try {
      const params = new URLSearchParams();
      if (filterCategory !== 'all') params.set('category', filterCategory);
      if (search) params.set('search', search);
      params.set('sort', sortOption);
      if (forceRefresh) params.set('refresh', '1');
      const res = await fetch(`/api/v1/skills?${params.toString()}`);
      if (!res.ok) throw new Error(`Failed to fetch skills: ${res.status}`);
      const data = await res.json();
      setSkills(data.skills || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load skills');
    } finally {
      setLoading(false);
    }
  }, [filterCategory, sortOption]);

  // Debounced search timer
  const searchTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  // AbortController for canceling in-flight requests
  const abortControllerRef = useRef<AbortController | null>(null);

  // Handle search input with debounce (triggers backend call)
  const handleSearchChange = useCallback((value: string) => {
    setSearchQuery(value);
    // Clear previous timer
    if (searchTimerRef.current) clearTimeout(searchTimerRef.current);
    // Cancel any in-flight request
    if (abortControllerRef.current) abortControllerRef.current.abort();
    // Debounce 300ms before fetching from backend
    searchTimerRef.current = setTimeout(() => {
      const controller = new AbortController();
      abortControllerRef.current = controller;
      fetchSkills(value);
    }, 300);
  }, [fetchSkills]);

  // Trigger fetch when filterCategory changes
  useEffect(() => {
    if (open) {
      fetchSkills(searchQuery);
    }
  }, [open, filterCategory]); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (open) {
      fetchSkills();
      fetchConfig();
    }
  }, [open, fetchSkills]);

  // Fetch current API key config
  const fetchConfig = useCallback(async () => {
    try {
      const res = await fetch('/api/v1/skills/config');
      if (res.ok) {
        const data = await res.json();
        setApiKey(data.apiKey || '');
      }
    } catch {
      // Silently ignore config errors
    }
  }, []);

  // Save API key
  const saveApiKey = useCallback(async () => {
    setSavingKey(true);
    setConfigStatus('saving');
    try {
      const res = await fetch('/api/v1/skills/config/apikey', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ apiKey: apiKeyInput }),
      });
      if (!res.ok) {
        throw new Error('Failed to save API key');
      }
      setApiKey(apiKeyInput);
      setConfigStatus('saved');
      setTimeout(() => setConfigStatus('idle'), 2000);
    } catch {
      setConfigStatus('error');
    } finally {
      setSavingKey(false);
    }
  }, [apiKeyInput]);

  // Install a skill
  const handleInstall = useCallback(async (skillId: string) => {
    setError(null);
    setInstallingId(skillId);
    try {
      const res = await fetch('/api/v1/skills/install', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ skillId }),
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || 'Install failed');
      }
      const data = await res.json();
      const installedIds = Array.isArray(data.installed) ? data.installed : [];
      setSkills((prev) =>
        prev.map((s) =>
          installedIds.includes(s.id) ? { ...s, installed: true } : s
        )
      );
    } catch (err) {
      console.error('Failed to install skill:', err);
      setError(err instanceof Error ? err.message : 'Install failed');
    } finally {
      setInstallingId(null);
    }
  }, []);

  // Cleanup: clear search timer on unmount
  useEffect(() => {
    return () => {
      if (searchTimerRef.current) clearTimeout(searchTimerRef.current);
      if (abortControllerRef.current) abortControllerRef.current.abort();
    };
  }, []);

  // Close on overlay click (but not panel click)
  const handleOverlayClick = useCallback((e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onOpenChange(false);
    }
  }, [onOpenChange]);

  // Handle Escape key
  useEffect(() => {
    if (!open) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onOpenChange(false);
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [open, onOpenChange]);

  // Get unique categories
  const categories = useMemo(
    () => ['all', ...Array.from(new Set(skills.map((s) => s.category)))],
    [skills]
  );

  // Sort skills (backend handles search and category filtering)
  const filteredSkills = useMemo(() => {
    let result = [...skills].sort((a, b) => {
      switch (sortOption) {
        case 'name':
          return a.name.localeCompare(b.name);
        case 'rating':
          return b.rating - a.rating;
        case 'installs':
          return b.downloads - a.downloads;
        default:
          return 0;
      }
    });
    return result;
  }, [skills, sortOption]);

  if (!open) return null;

  return (
    <div
      className="skill-marketplace-overlay"
      role="dialog"
      aria-modal="true"
      aria-label="Skill Marketplace"
      onClick={handleOverlayClick}
    >
      <div className="skill-marketplace-panel">
        {/* Header */}
        <div className="skill-marketplace-header">
          <div className="skill-marketplace-title-row">
            <Package size={20} className="skill-marketplace-icon" aria-hidden="true" />
            <h2>Skill Marketplace</h2>
          </div>
          <div className="skill-marketplace-header-actions">
            <button
              className="skill-marketplace-settings"
              onClick={() => {
                setShowSettings(true);
                setApiKeyInput(apiKey);
                setConfigStatus('idle');
              }}
              aria-label="Settings"
            >
              <Settings size={18} aria-hidden="true" />
            </button>
            <button
              className="skill-marketplace-close"
              onClick={() => onOpenChange(false)}
              aria-label="Close marketplace"
            >
              <X size={20} aria-hidden="true" />
            </button>
          </div>
        </div>

        {/* Settings Modal */}
        {showSettings && (
          <div className="skill-settings-overlay" onClick={() => setShowSettings(false)}>
            <div className="skill-settings-modal" onClick={(e) => e.stopPropagation()}>
              <div className="skill-settings-header">
                <h3>SkillsMP API Key</h3>
                <button onClick={() => setShowSettings(false)} aria-label="Close settings">
                  <X size={18} aria-hidden="true" />
                </button>
              </div>
              <p className="skill-settings-description">
                Get your API key from{' '}
                <a href="https://skillsmp.com/docs/api" target="_blank" rel="noopener noreferrer">
                  skillsmp.com/docs/api
                </a>
                {' '}to access the full skill registry.
              </p>
              <div className="skill-settings-input-row">
                <div className="skill-api-key-input">
                  <input
                    type={showKey ? 'text' : 'password'}
                    value={apiKeyInput}
                    onChange={(e) => setApiKeyInput(e.target.value)}
                    placeholder="sk_live_skillsmp_..."
                    aria-label="API Key"
                  />
                  <button
                    className="skill-api-key-toggle"
                    onClick={() => setShowKey(!showKey)}
                    aria-label={showKey ? 'Hide API key' : 'Show API key'}
                  >
                    {showKey ? <EyeOff size={16} /> : <Eye size={16} />}
                  </button>
                </div>
                <button
                  className="skill-settings-save"
                  onClick={saveApiKey}
                  disabled={savingKey || !apiKeyInput.trim()}
                >
                  {savingKey ? <Loader2 size={14} className="spin" /> : 'Save'}
                </button>
              </div>
              {configStatus === 'saved' && (
                <p className="skill-settings-status success">
                  <CheckCircle2 size={14} /> API key saved successfully
                </p>
              )}
              {configStatus === 'error' && (
                <p className="skill-settings-status error">
                  <AlertTriangle size={14} /> Failed to save API key
                </p>
              )}
            </div>
          </div>
        )}

        {/* Filters */}
        <div className="skill-marketplace-filters">
          <div className="skill-marketplace-search-row">
            <Search size={16} className="skill-marketplace-search-icon" aria-hidden="true" />
            <input
              type="text"
              className="skill-marketplace-search"
              placeholder="Search skills..."
              value={searchQuery}
              onChange={(e) => handleSearchChange(e.target.value)}
              aria-label="Search skills"
            />
            <button
              className="skill-marketplace-refresh"
              onClick={() => fetchSkills(searchQuery, true)}
              disabled={loading}
              aria-label="Refresh skills"
              title="Refresh from registry"
            >
              <RefreshCw size={16} className={loading ? 'spin' : ''} aria-hidden="true" />
            </button>
          </div>
          <div className="skill-marketplace-category-filter" role="group" aria-label="Filter by category">
            {categories.map((cat) => (
              <button
                key={cat}
                className={`category-btn ${filterCategory === cat ? 'active' : ''}`}
                onClick={() => setFilterCategory(cat)}
                aria-pressed={filterCategory === cat}
              >
                {cat === 'all' ? 'All' : cat}
              </button>
            ))}
          </div>
          <div className="skill-marketplace-sort-row">
            <label htmlFor="skill-sort" className="sr-only">Sort skills by</label>
            <select
              id="skill-sort"
              className="skill-marketplace-sort"
              value={sortOption}
              onChange={(e) => setSortOption(e.target.value as SortOption)}
            >
              <option value="name">Name</option>
              <option value="rating">Rating</option>
              <option value="installs">Installs</option>
            </select>
          </div>
        </div>

        {/* Loading state - skeleton grid */}
        {loading && (
          <div className="skill-marketplace-grid">
            {Array.from({ length: 6 }).map((_, i) => (
              <SkillCardSkeleton key={i} />
            ))}
          </div>
        )}

        {/* Error state */}
        {error && !loading && (
          <div className="skill-marketplace-error" role="alert">
            <AlertTriangle size={24} aria-hidden="true" />
            <div>
              <h3>Failed to load skills</h3>
              <p>{error}</p>
              <button className="skill-marketplace-retry" onClick={() => fetchSkills(searchQuery)}>
                Try Again
              </button>
            </div>
          </div>
        )}

        {/* Empty state */}
        {!loading && !error && skills.length === 0 && (
          <div className="skill-marketplace-empty" role="status">
            <Package size={32} aria-hidden="true" />
            <h3>No Skills Available</h3>
            <p>Browse the marketplace to discover and install skills</p>
          </div>
        )}

        {/* Skill grid */}
        {!loading && !error && skills.length > 0 && (
          <div className="skill-marketplace-grid">
            {filteredSkills.length === 0 ? (
              <div className="skill-marketplace-empty" role="status">
                <Search size={32} aria-hidden="true" />
                <h3>No results found</h3>
                <p>Try adjusting your search or filters</p>
              </div>
            ) : (
              filteredSkills.map((skill) => (
                <div key={skill.id} className="skill-card">
                  <div className="skill-card-header">
                    <div className="skill-icon" aria-hidden="true">
                      {(skill.icon || '?').charAt(0).toUpperCase()}
                    </div>
                    <div className="skill-card-title">
                      <h3>{skill.name}</h3>
                      <span className="skill-version">v{skill.version}</span>
                    </div>
                  </div>
                  <p className="skill-description">{skill.description}</p>
                  <div className="skill-meta">
                    <StarRating rating={skill.rating} />
                    <span className="skill-rating-count">({skill.ratingCount})</span>
                  </div>
                  <div className="skill-tags">
                    {(skill.tags || []).map((tag) => (
                      <span key={tag} className="skill-tag">
                        {tag}
                      </span>
                    ))}
                  </div>
                  <div className="skill-card-footer">
                    <div className="skill-footer-left">
                      <span className="skill-author">by {skill.author}</span>
                      <span className="skill-downloads">
                        <Download size={12} aria-hidden="true" />
                        {skill.downloads.toLocaleString()}
                      </span>
                    </div>
                    {skill.installed ? (
                      <span className="skill-installed-badge">
                        <CheckCircle2 size={14} aria-hidden="true" />
                        Installed
                      </span>
                    ) : (
                      <button
                        className="skill-install-btn"
                        onClick={() => handleInstall(skill.id)}
                        disabled={installingId !== null}
                        aria-label={`Install ${skill.name}`}
                      >
                        {installingId === skill.id ? (
                          <>
                            <Loader2 size={14} className="spin" aria-hidden="true" />
                            Installing...
                          </>
                        ) : (
                          'Install'
                        )}
                      </button>
                    )}
                  </div>
                </div>
              ))
            )}
          </div>
        )}
      </div>
    </div>
  );
};
