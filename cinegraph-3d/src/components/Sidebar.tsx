import React, { useState, useEffect, useMemo } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { X, Film, User, Tag, Star, Calendar, Clock, Shield, Bookmark, BookmarkCheck, CheckCircle2, Clapperboard, Users, TrendingUp, Sparkles } from 'lucide-react';
import { Node, GraphData, Link, api, WatchlistEntry } from '../services/api';

interface SidebarProps {
  node: Node | null;
  onClose: () => void;
  watchlist: WatchlistEntry[];
  onUpdateWatchlist: () => void;
  graphData: GraphData;
}

// --- Helper: derive connected nodes from graph ---
function getConnectedNodes(nodeId: string, graphData: GraphData, linkLabel?: string) {
  const connected: { node: Node; label: string }[] = [];
  for (const link of graphData.links) {
    const sourceId = typeof link.source === 'object' ? (link.source as any).id : link.source;
    const targetId = typeof link.target === 'object' ? (link.target as any).id : link.target;

    if (linkLabel && link.label !== linkLabel) continue;

    let otherId: string | null = null;
    if (sourceId === nodeId) otherId = targetId;
    else if (targetId === nodeId) otherId = sourceId;

    if (otherId) {
      const otherNode = graphData.nodes.find(n => n.id === otherId);
      if (otherNode) connected.push({ node: otherNode, label: link.label });
    }
  }
  return connected;
}

// --- Sub-components ---

function FilmMiniCard({ film, rank }: { film: Node; rank?: number }) {
  return (
    <div className="flex items-center gap-3 bg-white/5 hover:bg-white/10 rounded-lg p-3 border border-white/5 transition-colors">
      {rank && (
        <span className="text-lg font-bold text-white/20 w-6 text-center font-mono">{rank}</span>
      )}
      {film.image && (
        <img src={film.image} alt={film.name} className="w-10 h-14 rounded object-cover flex-shrink-0" />
      )}
      <div className="min-w-0 flex-1">
        <div className="text-sm font-semibold text-white/90 truncate">{film.name}</div>
        <div className="flex items-center gap-2 text-xs text-white/50">
          {film.year && <span>{film.year}</span>}
          {film.rating && (
            <span className="flex items-center gap-0.5 text-yellow-500">
              <Star size={10} className="fill-current" /> {film.rating}
            </span>
          )}
        </div>
      </div>
    </div>
  );
}

function SectionHeader({ icon: Icon, title, color }: { icon: any; title: string; color: string }) {
  return (
    <div className="flex items-center gap-2 mb-3">
      <Icon size={14} className={color} />
      <h3 className="text-xs font-mono uppercase tracking-widest text-white/40">{title}</h3>
    </div>
  );
}

function TagPill({ label, color }: { label: string; color?: string }) {
  return (
    <span className={`inline-block px-2.5 py-1 rounded-full text-xs font-medium border ${color || 'bg-white/5 border-white/10 text-white/70'}`}>
      {label}
    </span>
  );
}

// =============================================================================
// Movie Detail Panel
// =============================================================================
function MoviePanel({
  node, graphData, watchlist, onUpdateWatchlist,
}: {
  node: Node; graphData: GraphData; watchlist: WatchlistEntry[]; onUpdateWatchlist: () => void;
}) {
  const [isEditingWatchlist, setIsEditingWatchlist] = useState(false);
  const [notes, setNotes] = useState('');
  const [priority, setPriority] = useState(5);
  const [rating, setRating] = useState<number | ''>('');

  const watchlistEntry = watchlist.find(w => w.film_id === node.id);
  const isInWatchlist = !!watchlistEntry;

  useEffect(() => {
    if (watchlistEntry) {
      setNotes(watchlistEntry.notes || '');
      setPriority(watchlistEntry.priority || 5);
      setRating(watchlistEntry.rating || '');
    } else {
      setNotes('');
      setPriority(5);
      setRating('');
    }
    setIsEditingWatchlist(false);
  }, [node, watchlistEntry]);

  const connected = useMemo(() => getConnectedNodes(node.id, graphData), [node.id, graphData]);
  const actors = connected.filter(c => c.label === 'ACTED_IN' && c.node.type === 'Person');
  const directors = connected.filter(c => c.label === 'DIRECTED' && c.node.type === 'Person');
  const genres = connected.filter(c => c.label === 'IN_GENRE' && c.node.type === 'Genre');

  const handleSave = async () => {
    if (rating !== '') {
      await api.rateFilm(node.id, Number(rating), notes, priority);
    } else {
      await api.addToWatchlist(node.id, notes, priority);
    }
    onUpdateWatchlist();
    setIsEditingWatchlist(false);
  };

  const handleRemove = async () => {
    if (!watchlistEntry) return;
    await api.removeFromWatchlist(node.id, watchlistEntry.id);
    onUpdateWatchlist();
    setIsEditingWatchlist(false);
  };

  return (
    <>
      {/* Poster */}
      {node.image && (
        <div className="mb-5 rounded-xl overflow-hidden border border-white/10">
          <img src={node.image} alt={node.name} className="w-full h-48 object-cover" />
        </div>
      )}

      {/* Metadata Row */}
      <div className="flex flex-wrap gap-3 mb-5">
        {node.year && (
          <div className="flex items-center gap-1.5 text-white/70 text-sm">
            <Calendar size={14} /> {node.year}
          </div>
        )}
        {node.rating && (
          <div className="flex items-center gap-1.5 text-yellow-500 text-sm">
            <Star size={14} className="fill-current" /> {node.rating}/10
          </div>
        )}
        {node.runtime && (
          <div className="flex items-center gap-1.5 text-white/70 text-sm">
            <Clock size={14} /> {node.runtime}
          </div>
        )}
        {node.certificate && (
          <div className="flex items-center gap-1.5 text-white/70 text-sm">
            <Shield size={14} /> {node.certificate}
          </div>
        )}
      </div>

      {/* Description */}
      {node.description && (
        <div className="mb-6">
          <p className="text-white/70 leading-relaxed text-sm">{node.description}</p>
        </div>
      )}

      {/* Genres */}
      {genres.length > 0 && (
        <div className="mb-5">
          <SectionHeader icon={Tag} title="Genres" color="text-emerald-500" />
          <div className="flex flex-wrap gap-2">
            {genres.map(g => (
              <TagPill key={g.node.id} label={g.node.name} color="bg-emerald-500/10 border-emerald-500/30 text-emerald-400" />
            ))}
          </div>
        </div>
      )}

      {/* Directors */}
      {directors.length > 0 && (
        <div className="mb-5">
          <SectionHeader icon={Clapperboard} title="Directors" color="text-purple-400" />
          <div className="flex flex-wrap gap-2">
            {directors.map(d => (
              <TagPill key={d.node.id} label={d.node.name} color="bg-purple-500/10 border-purple-500/30 text-purple-400" />
            ))}
          </div>
        </div>
      )}

      {/* Actors */}
      {actors.length > 0 && (
        <div className="mb-6">
          <SectionHeader icon={Users} title="Cast" color="text-sky-400" />
          <div className="flex flex-wrap gap-2">
            {actors.map(a => (
              <TagPill key={a.node.id} label={a.node.name} color="bg-sky-500/10 border-sky-500/30 text-sky-400" />
            ))}
          </div>
        </div>
      )}

      {/* Watchlist */}
      <div className="bg-white/5 border border-white/10 rounded-xl p-4">
        {!isEditingWatchlist ? (
          <div>
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                {isInWatchlist ? (
                  watchlistEntry?.watched ? <CheckCircle2 size={18} className="text-emerald-500" /> : <BookmarkCheck size={18} className="text-[#d4a359]" />
                ) : (
                  <Bookmark size={18} className="text-white/50" />
                )}
                <span className="font-semibold text-sm">
                  {isInWatchlist
                    ? (watchlistEntry?.watched ? 'Watched' : 'In Watchlist')
                    : 'Not in Watchlist'}
                </span>
              </div>
              <button
                onClick={() => setIsEditingWatchlist(true)}
                className="text-xs bg-white/10 hover:bg-white/20 px-3 py-1.5 rounded transition-colors"
              >
                {isInWatchlist ? 'Edit' : 'Add'}
              </button>
            </div>
            {isInWatchlist && (
              <div className="mt-3 space-y-2 text-sm text-white/70">
                {watchlistEntry?.rating && (
                  <div className="flex items-center gap-2">
                    <span className="text-white/40 w-16">My Rating:</span>
                    <span className="text-yellow-500 flex items-center gap-1">
                      <Star size={12} className="fill-current" /> {watchlistEntry.rating}/10
                    </span>
                  </div>
                )}
                <div className="flex items-center gap-2">
                  <span className="text-white/40 w-16">Priority:</span>
                  <span>{watchlistEntry?.priority}/10</span>
                </div>
                {watchlistEntry?.notes && (
                  <div className="flex gap-2">
                    <span className="text-white/40 w-16">Notes:</span>
                    <span className="italic">"{watchlistEntry.notes}"</span>
                  </div>
                )}
              </div>
            )}
          </div>
        ) : (
          <div className="space-y-4">
            <div>
              <label className="text-xs text-white/50 mb-1 block">Priority (1-10)</label>
              <input
                type="number" min="1" max="10"
                value={priority} onChange={e => setPriority(Number(e.target.value))}
                className="w-full bg-[#1a1a1a] border border-white/10 rounded px-3 py-2 text-sm focus:outline-none focus:border-[#d4a359]"
              />
            </div>
            <div>
              <label className="text-xs text-white/50 mb-1 block">My Rating (1-10, optional)</label>
              <input
                type="number" min="1" max="10"
                value={rating} onChange={e => setRating(e.target.value ? Number(e.target.value) : '')}
                placeholder="Leave empty if not watched"
                className="w-full bg-[#1a1a1a] border border-white/10 rounded px-3 py-2 text-sm focus:outline-none focus:border-[#d4a359]"
              />
            </div>
            <div>
              <label className="text-xs text-white/50 mb-1 block">Notes</label>
              <textarea
                value={notes} onChange={e => setNotes(e.target.value)}
                className="w-full bg-[#1a1a1a] border border-white/10 rounded px-3 py-2 text-sm focus:outline-none focus:border-[#d4a359] min-h-[80px]"
              />
            </div>
            <div className="flex gap-2 pt-2">
              <button onClick={handleSave} className="flex-1 bg-[#d4a359] hover:bg-[#e5b46a] text-black font-medium py-2 rounded text-sm transition-colors">
                Save
              </button>
              {isInWatchlist && (
                <button onClick={handleRemove} className="flex-1 bg-red-500/20 hover:bg-red-500/30 text-red-400 font-medium py-2 rounded text-sm transition-colors border border-red-500/30">
                  Remove
                </button>
              )}
              <button onClick={() => setIsEditingWatchlist(false)} className="flex-1 bg-white/10 hover:bg-white/20 py-2 rounded text-sm transition-colors">
                Cancel
              </button>
            </div>
          </div>
        )}
      </div>
    </>
  );
}

// =============================================================================
// Person Detail Panel (Actor / Director)
// =============================================================================
function PersonPanel({ node, graphData }: { node: Node; graphData: GraphData }) {
  const connected = useMemo(() => getConnectedNodes(node.id, graphData), [node.id, graphData]);

  const actedIn = connected.filter(c => c.label === 'ACTED_IN' && c.node.type === 'Movie');
  const directed = connected.filter(c => c.label === 'DIRECTED' && c.node.type === 'Movie');

  const allFilms = [...actedIn.map(c => c.node), ...directed.map(c => c.node)];
  const uniqueFilms = Array.from(new Map(allFilms.map(f => [f.id, f])).values());
  const topFilms = [...uniqueFilms].sort((a, b) => (b.rating || 0) - (a.rating || 0)).slice(0, 3);

  // Determine role
  const isActor = actedIn.length > 0;
  const isDirector = directed.length > 0;
  const roleLabel = isActor && isDirector ? 'Actor & Director' : isDirector ? 'Director' : 'Actor';

  // Genre tendencies: find genres connected to this person's films
  const genreCounts = new Map<string, number>();
  for (const film of uniqueFilms) {
    const filmLinks = getConnectedNodes(film.id, graphData, 'IN_GENRE');
    for (const g of filmLinks) {
      genreCounts.set(g.node.name, (genreCounts.get(g.node.name) || 0) + 1);
    }
  }
  const topGenres = [...genreCounts.entries()].sort((a, b) => b[1] - a[1]).slice(0, 5);

  // Frequent collaborator: find other Person nodes who share the most films
  const collabCounts = new Map<string, { name: string; count: number }>();
  for (const film of uniqueFilms) {
    const filmPeople = getConnectedNodes(film.id, graphData);
    for (const p of filmPeople) {
      if (p.node.type === 'Person' && p.node.id !== node.id) {
        const existing = collabCounts.get(p.node.id);
        if (existing) existing.count++;
        else collabCounts.set(p.node.id, { name: p.node.name, count: 1 });
      }
    }
  }
  const topCollaborator = [...collabCounts.values()].sort((a, b) => b.count - a.count)[0];

  // Average rating
  const ratedFilms = uniqueFilms.filter(f => f.rating);
  const avgRating = ratedFilms.length > 0
    ? (ratedFilms.reduce((sum, f) => sum + (f.rating || 0), 0) / ratedFilms.length).toFixed(1)
    : null;

  return (
    <>
      {/* Role Badge */}
      <div className="mb-5">
        <TagPill label={roleLabel} color="bg-sky-500/10 border-sky-500/30 text-sky-400" />
      </div>

      {/* Stats Row */}
      <div className="grid grid-cols-3 gap-3 mb-6">
        {isActor && (
          <div className="bg-white/5 rounded-xl p-3 border border-white/5 text-center">
            <div className="text-xl font-light text-white/90">{actedIn.length}</div>
            <div className="text-[10px] text-white/40 uppercase tracking-wider">Acted In</div>
          </div>
        )}
        {isDirector && (
          <div className="bg-white/5 rounded-xl p-3 border border-white/5 text-center">
            <div className="text-xl font-light text-white/90">{directed.length}</div>
            <div className="text-[10px] text-white/40 uppercase tracking-wider">Directed</div>
          </div>
        )}
        {avgRating && (
          <div className="bg-white/5 rounded-xl p-3 border border-white/5 text-center">
            <div className="text-xl font-light text-yellow-500">{avgRating}</div>
            <div className="text-[10px] text-white/40 uppercase tracking-wider">Avg Rating</div>
          </div>
        )}
      </div>

      {/* Top Films */}
      {topFilms.length > 0 && (
        <div className="mb-6">
          <SectionHeader icon={TrendingUp} title="Top Films" color="text-rose-400" />
          <div className="space-y-2">
            {topFilms.map((f, i) => <FilmMiniCard key={f.id} film={f} rank={i + 1} />)}
          </div>
        </div>
      )}

      {/* Genre Tendencies */}
      {topGenres.length > 0 && (
        <div className="mb-6">
          <SectionHeader icon={Tag} title="Genre Tendencies" color="text-emerald-500" />
          <div className="space-y-2">
            {topGenres.map(([genre, count]) => (
              <div key={genre} className="flex items-center justify-between text-sm">
                <span className="text-white/70">{genre}</span>
                <span className="text-white/30 text-xs font-mono">{count} {count === 1 ? 'film' : 'films'}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Frequent Collaborator */}
      {topCollaborator && topCollaborator.count >= 2 && (
        <div className="bg-white/5 border border-white/10 rounded-xl p-4">
          <SectionHeader icon={Sparkles} title="Frequent Collaborator" color="text-[#d4a359]" />
          <div className="flex items-center justify-between">
            <span className="text-white/90 font-medium">{topCollaborator.name}</span>
            <span className="text-white/40 text-xs font-mono">{topCollaborator.count} films together</span>
          </div>
        </div>
      )}
    </>
  );
}

// =============================================================================
// Genre Detail Panel
// =============================================================================
function GenrePanel({ node, graphData }: { node: Node; graphData: GraphData }) {
  const connected = useMemo(() => getConnectedNodes(node.id, graphData, 'IN_GENRE'), [node.id, graphData]);
  const films = connected.filter(c => c.node.type === 'Movie').map(c => c.node);
  const topFilms = [...films].sort((a, b) => (b.rating || 0) - (a.rating || 0)).slice(0, 3);

  // Average rating
  const ratedFilms = films.filter(f => f.rating);
  const avgRating = ratedFilms.length > 0
    ? (ratedFilms.reduce((sum, f) => sum + (f.rating || 0), 0) / ratedFilms.length).toFixed(1)
    : null;

  // Common directors in this genre
  const directorCounts = new Map<string, { name: string; count: number }>();
  for (const film of films) {
    const filmPeople = getConnectedNodes(film.id, graphData, 'DIRECTED');
    for (const p of filmPeople) {
      if (p.node.type === 'Person') {
        const existing = directorCounts.get(p.node.id);
        if (existing) existing.count++;
        else directorCounts.set(p.node.id, { name: p.node.name, count: 1 });
      }
    }
  }
  const topDirectors = [...directorCounts.values()].sort((a, b) => b.count - a.count).slice(0, 3);

  // Common actors in this genre
  const actorCounts = new Map<string, { name: string; count: number }>();
  for (const film of films) {
    const filmPeople = getConnectedNodes(film.id, graphData, 'ACTED_IN');
    for (const p of filmPeople) {
      if (p.node.type === 'Person') {
        const existing = actorCounts.get(p.node.id);
        if (existing) existing.count++;
        else actorCounts.set(p.node.id, { name: p.node.name, count: 1 });
      }
    }
  }
  const topActors = [...actorCounts.values()].sort((a, b) => b.count - a.count).slice(0, 3);

  return (
    <>
      {/* Stats */}
      <div className="grid grid-cols-2 gap-3 mb-6">
        <div className="bg-white/5 rounded-xl p-3 border border-white/5 text-center">
          <div className="text-2xl font-light text-white/90">{films.length}</div>
          <div className="text-[10px] text-white/40 uppercase tracking-wider">Films</div>
        </div>
        {avgRating && (
          <div className="bg-white/5 rounded-xl p-3 border border-white/5 text-center">
            <div className="text-2xl font-light text-yellow-500">{avgRating}</div>
            <div className="text-[10px] text-white/40 uppercase tracking-wider">Avg Rating</div>
          </div>
        )}
      </div>

      {/* Top Films */}
      {topFilms.length > 0 && (
        <div className="mb-6">
          <SectionHeader icon={TrendingUp} title="Top Films" color="text-rose-400" />
          <div className="space-y-2">
            {topFilms.map((f, i) => <FilmMiniCard key={f.id} film={f} rank={i + 1} />)}
          </div>
        </div>
      )}

      {/* Top Directors */}
      {topDirectors.length > 0 && (
        <div className="mb-6">
          <SectionHeader icon={Clapperboard} title="Top Directors" color="text-purple-400" />
          <div className="space-y-2">
            {topDirectors.map(d => (
              <div key={d.name} className="flex items-center justify-between text-sm">
                <span className="text-white/70">{d.name}</span>
                <span className="text-white/30 text-xs font-mono">{d.count} {d.count === 1 ? 'film' : 'films'}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Top Actors */}
      {topActors.length > 0 && (
        <div className="mb-6">
          <SectionHeader icon={Users} title="Top Actors" color="text-sky-400" />
          <div className="space-y-2">
            {topActors.map(a => (
              <div key={a.name} className="flex items-center justify-between text-sm">
                <span className="text-white/70">{a.name}</span>
                <span className="text-white/30 text-xs font-mono">{a.count} {a.count === 1 ? 'film' : 'films'}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </>
  );
}

// =============================================================================
// Main Sidebar
// =============================================================================
export default function Sidebar({ node, onClose, watchlist, onUpdateWatchlist, graphData }: SidebarProps) {
  return (
    <AnimatePresence>
      {node && (
        <motion.div
          key={node.id}
          initial={{ x: '100%', opacity: 0 }}
          animate={{ x: 0, opacity: 1 }}
          exit={{ x: '100%', opacity: 0 }}
          transition={{ type: 'spring', damping: 25, stiffness: 200 }}
          className="fixed right-0 top-0 bottom-0 w-80 md:w-96 bg-[#1a1a1a]/95 backdrop-blur-xl border-l border-white/5 p-6 z-50 text-white overflow-y-auto shadow-2xl"
        >
          <button
            onClick={onClose}
            className="absolute top-6 right-6 p-2 rounded-full hover:bg-white/10 transition-colors"
          >
            <X size={20} />
          </button>

          <div className="mt-8">
            {/* Type Badge + Name */}
            <div className="flex items-center gap-3 mb-2">
              {node.type === 'Movie' && <Film className="text-rose-500" size={24} />}
              {node.type === 'Person' && <User className="text-sky-500" size={24} />}
              {node.type === 'Genre' && <Tag className="text-emerald-500" size={24} />}
              <span className="text-sm font-mono tracking-wider uppercase text-white/50">
                {node.type}
              </span>
            </div>

            <h2 className="text-3xl font-bold font-serif mb-6 leading-tight text-white/90">
              {node.name}
            </h2>

            {/* Type-specific panels */}
            {node.type === 'Movie' && (
              <MoviePanel
                node={node}
                graphData={graphData}
                watchlist={watchlist}
                onUpdateWatchlist={onUpdateWatchlist}
              />
            )}

            {node.type === 'Person' && (
              <PersonPanel node={node} graphData={graphData} />
            )}

            {node.type === 'Genre' && (
              <GenrePanel node={node} graphData={graphData} />
            )}
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
