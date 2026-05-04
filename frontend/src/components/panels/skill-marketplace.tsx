import React, { useState, useEffect, useCallback } from 'react';

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

export const SkillMarketplace: React.FC<SkillMarketplaceProps> = ({ open, onOpenChange }) => {
  const [skills, setSkills] = useState<Skill[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filterCategory, setFilterCategory] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');

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
      setSkills((prev) =>
        prev.map((s) =>
          data.installed.includes(s.id) ? { ...s, installed: true } : s
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
  const categories = ['all', ...Array.from(new Set(skills.map((s) => s.category)))];

  // Filter by search query
  const filteredSkills = skills.filter((s) => {
    if (searchQuery === '') return true;
    const q = searchQuery.toLowerCase();
    return (
      s.name.toLowerCase().includes(q) ||
      s.description.toLowerCase().includes(q) ||
      s.tags.some((t) => t.toLowerCase().includes(q))
    );
  });

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
        </div>

        {loading && (
          <div className="skill-marketplace-loading" role="status" aria-live="polite">
            Loading skills...
          </div>
        )}

        {error && (
          <div className="skill-marketplace-error" role="alert">
            {error}
          </div>
        )}

        {!loading && !error && (
          <div className="skill-marketplace-grid">
            {filteredSkills.map((skill) => (
              <div key={skill.id} className="skill-card">
                <div className="skill-card-header">
                  <div className="skill-icon">{skill.icon.charAt(0).toUpperCase()}</div>
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
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
