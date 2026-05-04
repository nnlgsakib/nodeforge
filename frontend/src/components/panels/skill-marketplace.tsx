import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { EmptyState } from '../ui/EmptyState';

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

/** Renders a star rating display */
function StarRating({ rating }: { rating: number }) {
  const clamped = Math.max(0, Math.min(5, rating));
  const fullStars = Math.floor(clamped);
  const hasHalf = clamped - fullStars >= 0.5;
  const emptyStars = 5 - fullStars - (hasHalf ? 1 : 0);

  return (
    <span className="skill-stars" aria-label={`Rating: ${clamped} out of 5 stars`}>
      {Array.from({ length: fullStars }).map((_, i) => (
        <span key={`full-${i}`} className="star star-full">&#9733;</span>
      ))}
      {hasHalf && <span className="star star-half">&#9733;</span>}
      {Array.from({ length: Math.max(0, emptyStars) }).map((_, i) => (
        <span key={`empty-${i}`} className="star star-empty">&#9734;</span>
      ))}
    </span>
  );
}

type SortOption = 'name' | 'rating' | 'installs' | 'recent';

export const SkillMarketplace: React.FC<SkillMarketplaceProps> = ({ open, onOpenChange }) => {
  const [skills, setSkills] = useState<Skill[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filterCategory, setFilterCategory] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [sortOption, setSortOption] = useState<SortOption>('name');

  // Fetch skills from backend
  const fetchSkills = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const params = filterCategory !== 'all' ? `?category=${encodeURIComponent(filterCategory)}` : '';
      const res = await fetch(`/api/v1/skills${params}`);
      if (!res.ok) throw new Error(`Failed to fetch skills: ${res.status}`);
      const data = await res.json();
      setSkills(data.skills || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load skills');
    } finally {
      setLoading(false);
    }
  }, [filterCategory]);

  useEffect(() => {
    if (open) {
      fetchSkills();
    }
  }, [open, fetchSkills]);

  // Install a skill
  const handleInstall = useCallback(async (skillId: string) => {
    setError(null); // clear previous errors before retry
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
      // Update installed status for all installed skills
      const installedIds = Array.isArray(data.installed) ? data.installed : [];
      setSkills((prev) =>
        prev.map((s) =>
          installedIds.includes(s.id) ? { ...s, installed: true } : s
        )
      );
    } catch (err) {
      console.error('Failed to install skill:', err);
      setError(err instanceof Error ? err.message : 'Install failed');
    }
  }, []);

  // Close on overlay click (but not panel click)
  const handleOverlayClick = useCallback((e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onOpenChange(false);
    }
  }, [onOpenChange]);

  // Get unique categories
  const categories = useMemo(() => ['all', ...Array.from(new Set(skills.map((s) => s.category)))], [skills]);

  // Filter and sort skills
  const filteredSkills = useMemo(() => {
    let result = skills.filter((s) => {
      if (filterCategory !== 'all' && s.category !== filterCategory) return false;
      if (searchQuery === '') return true;
      const q = searchQuery.toLowerCase();
      return (
        (s.name ?? '').toLowerCase().includes(q) ||
        (s.description ?? '').toLowerCase().includes(q) ||
        s.tags.some((t) => t.toLowerCase().includes(q))
      );
    });

    // Apply sorting
    result = [...result].sort((a, b) => {
      switch (sortOption) {
        case 'name':
          return a.name.localeCompare(b.name);
        case 'rating':
          return b.rating - a.rating;
        case 'installs':
          return b.downloads - a.downloads;
        case 'recent':
          // Would need a createdAt field; fallback to id comparison
          return b.id.localeCompare(a.id);
        default:
          return 0;
      }
    });

    return result;
  }, [skills, searchQuery, sortOption, filterCategory]);

  if (!open) return null;

  return (
    <div className="skill-marketplace-overlay" role="dialog" aria-modal="true" aria-label="Skill Marketplace" onClick={handleOverlayClick}>
      <div className="skill-marketplace-panel">
        <div className="skill-marketplace-header">
          <h2>Skill Marketplace</h2>
          <button
            className="skill-marketplace-close"
            onClick={() => onOpenChange(false)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                onOpenChange(false);
              }
            }}
            aria-label="Close marketplace"
          >
            &times;
          </button>
        </div>

        <div className="skill-marketplace-filters">
          <input
            type="text"
            className="skill-marketplace-search"
            placeholder="Search skills..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            aria-label="Search skills"
          />
          <div className="skill-marketplace-category-filter">
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
          <div className="skill-marketplace-sort">
            <label htmlFor="skill-sort" className="sr-only">Sort skills by</label>
            <select
              id="skill-sort"
              value={sortOption}
              onChange={(e) => setSortOption(e.target.value as SortOption)}
              style={{
                padding: '6px 8px',
                border: '1px solid var(--bg-tertiary)',
                borderRadius: '6px',
                background: 'var(--bg-primary)',
                color: 'var(--text-secondary)',
                fontSize: '12px',
                cursor: 'pointer',
              }}
            >
              <option value="name">Sort: Name</option>
              <option value="rating">Sort: Rating</option>
              <option value="installs">Sort: Installs</option>
              <option value="recent">Sort: Recent</option>
            </select>
          </div>
        </div>

        {loading && (
          <EmptyState icon={<span aria-hidden="true">⏳</span>} title="Loading skills..." animated />
        )}

        {error && (
          <EmptyState
            icon={<span aria-hidden="true">⚠️</span>}
            title="Failed to load skills"
            description={error}
          />
        )}

        {!loading && !error && skills.length === 0 && (
          <EmptyState
            icon={<span aria-hidden="true">🔌</span>}
            title="No Skills Installed"
            description="Browse the marketplace to discover and install skills"
          />
        )}

        {!loading && !error && skills.length > 0 && (
          <div className="skill-marketplace-grid">
            {filteredSkills.length === 0 ? (
              <EmptyState
                icon={<span aria-hidden="true">🕭</span>}
                title="No skills match your search"
                description="Try adjusting your filters or search terms"
              />
            ) : (
              filteredSkills.map((skill) => (
              <div key={skill.id} className="skill-card">
                <div className="skill-card-header">
                  <div className="skill-icon">{(skill.icon || '?').charAt(0).toUpperCase()}</div>
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
                  {skill.tags.map((tag) => (
                    <span key={tag} className="skill-tag">{tag}</span>
                  ))}
                </div>
                <div className="skill-card-footer">
                  <span className="skill-author">by {skill.author}</span>
                  <span className="skill-downloads">{skill.downloads.toLocaleString()} installs</span>
                  {skill.installed ? (
                    <span className="skill-installed-badge">Installed</span>
                  ) : (
                    <button
                      className="skill-install-btn"
                      onClick={() => handleInstall(skill.id)}
                      aria-label={`Install ${skill.name}`}
                    >
                      Install
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
