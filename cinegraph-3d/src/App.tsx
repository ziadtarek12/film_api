import React, { useState, useMemo, useEffect, useRef } from 'react';
import GraphView, { GraphViewHandle } from './components/GraphView';
import Sidebar from './components/Sidebar';
import SearchBar from './components/SearchBar';
import ControlsPanel from './components/ControlsPanel';
import { api, Node, GraphData, WatchlistEntry } from './services/api';
import { motion, AnimatePresence } from 'motion/react';
import { Bookmark, Sparkles, Loader2, Wand2 } from 'lucide-react';

export interface FilterGroup {
  id: string;
  query: string;
  color: string;
  hidden?: boolean;
}

export default function App() {
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const [watchlist, setWatchlist] = useState<WatchlistEntry[]>([]);
  const [recommendations, setRecommendations] = useState<Node[]>([]);
  const [graphMode, setGraphMode] = useState<'all' | 'watchlist' | 'recommendations'>('all');
  const [graphData, setGraphData] = useState<GraphData>({ nodes: [], links: [] });
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const graphViewRef = useRef<GraphViewHandle>(null);

  const [forces, setForces] = useState({
    centerForce: 0.5,
    repelForce: 100,
    linkForce: 0.5,
    linkDistance: 30,
  });

  const [filters, setFilters] = useState({
    title: '',
    genres: '',
    actors: '',
    directors: '',
  });

  const [display, setDisplay] = useState({
    showLabels: false,
    nodeSize: 'influence' as const,
    linkThickness: 0.5,
  });

  const [groups, setGroups] = useState<FilterGroup[]>([
    { id: '1', query: 'type:Movie', color: '#9333ea', hidden: false },
    { id: '2', query: 'type:Person', color: '#0ea5e9', hidden: false },
    { id: '3', query: 'type:Genre', color: '#10b981', hidden: false },
  ]);

  useEffect(() => {
    setIsLoading(true);
    setError(null);
    Promise.all([
      api.getFilmsAsGraph(100),
      api.getWatchlist().catch(() => []),
      api.getRecommendations().catch(() => [])
    ]).then(([graph, wList, recs]) => {
      if (graph.nodes.length === 0) {
        setError('The API returned no films. Please check if your database is populated.');
      }
      setGraphData(graph);
      setWatchlist(wList);
      setRecommendations(recs);
    }).catch(err => {
      console.error('API Error:', err);
      setError(err.message || 'Failed to fetch data from the API.');
    }).finally(() => {
      setIsLoading(false);
    });
  }, []);

  const handleNodeClick = (node: Node) => {
    setSelectedNode(node);
  };

  const handleCloseSidebar = () => {
    setSelectedNode(null);
  };

  const handleUpdateWatchlist = async () => {
    const updated = await api.getWatchlist();
    setWatchlist(updated);
  };

  // Filter logic
  const filteredData = useMemo(() => {
    if (!graphData.nodes.length) return { nodes: [], links: [] };

    let baseMovies = graphData.nodes.filter(n => n.type === 'Movie');

    // Apply graph mode (Watchlist / For You)
    if (graphMode === 'watchlist') {
      const watchlistIds = new Set(watchlist.map(w => w.film_id));
      baseMovies = baseMovies.filter(m => watchlistIds.has(m.id));
    } else if (graphMode === 'recommendations') {
      const recIds = new Set(recommendations.map(r => r.id));
      baseMovies = baseMovies.filter(m => recIds.has(m.id));
    }

    const titleQ = filters.title.toLowerCase();
    const genresQ = filters.genres.toLowerCase().split(',').map(s => s.trim()).filter(Boolean);
    const actorsQ = filters.actors.toLowerCase().split(',').map(s => s.trim()).filter(Boolean);
    const directorsQ = filters.directors.toLowerCase().split(',').map(s => s.trim()).filter(Boolean);

    // Find movies that match the text criteria
    const matchingMovies = new Set<string>();

    baseMovies.forEach(node => {
      let matches = true;
      if (titleQ && !node.name.toLowerCase().includes(titleQ)) matches = false;

      // Get connected nodes for this movie
      const connectedNodes = graphData.links
        .filter(l => {
          const sourceId = typeof l.source === 'object' ? (l.source as any).id : l.source;
          const targetId = typeof l.target === 'object' ? (l.target as any).id : l.target;
          return sourceId === node.id || targetId === node.id;
        })
        .map(l => {
          const sourceId = typeof l.source === 'object' ? (l.source as any).id : l.source;
          const targetId = typeof l.target === 'object' ? (l.target as any).id : l.target;
          const otherId = sourceId === node.id ? targetId : sourceId;
          return graphData.nodes.find(n => n.id === otherId);
        })
        .filter(Boolean) as Node[];

      if (genresQ.length > 0) {
        const movieGenres = connectedNodes.filter(n => n.type === 'Genre').map(n => n.name.toLowerCase());
        if (!genresQ.some(g => movieGenres.some(mg => mg.includes(g)))) matches = false;
      }

      if (actorsQ.length > 0) {
        const movieActors = connectedNodes.filter(n => n.type === 'Person').map(n => n.name.toLowerCase());
        if (!actorsQ.some(a => movieActors.some(ma => ma.includes(a)))) matches = false;
      }

      if (directorsQ.length > 0) {
        const movieDirectors = connectedNodes.filter(n => n.type === 'Person').map(n => n.name.toLowerCase());
        if (!directorsQ.some(d => movieDirectors.some(md => md.includes(d)))) matches = false;
      }

      if (matches) {
        matchingMovies.add(node.id);
      }
    });

    // Now build the graph with only matching movies and their immediate connections
    const newNodes = new Set<Node>();
    const newLinks = graphData.links.filter(l => {
      const sourceId = typeof l.source === 'object' ? (l.source as any).id : l.source;
      const targetId = typeof l.target === 'object' ? (l.target as any).id : l.target;

      if (matchingMovies.has(sourceId) || matchingMovies.has(targetId)) {
        const sourceNode = graphData.nodes.find(n => n.id === sourceId);
        const targetNode = graphData.nodes.find(n => n.id === targetId);
        if (sourceNode) newNodes.add(sourceNode);
        if (targetNode) newNodes.add(targetNode);
        return true;
      }
      return false;
    });

    return {
      nodes: Array.from(newNodes),
      links: newLinks
    };
  }, [filters, graphMode, watchlist, recommendations, groups, graphData]);

  if (isLoading) {
    return (
      <div className="w-full h-screen bg-[#000000] flex items-center justify-center text-white/50 flex-col gap-4">
        <Loader2 className="animate-spin text-[#d4a359]" size={32} />
        <p className="font-mono text-sm tracking-widest uppercase text-center px-4">Connecting to Database...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="w-full h-screen bg-[#000000] flex items-center justify-center text-white flex-col gap-6 p-8 text-center">
        <div className="w-16 h-16 rounded-full bg-red-500/20 flex items-center justify-center border border-red-500/50">
          <span className="text-red-500 text-2xl font-bold">!</span>
        </div>
        <div>
          <h2 className="text-2xl font-bold mb-2">Connection Issue</h2>
          <p className="text-white/60 max-w-md mx-auto font-mono text-sm">{error}</p>
        </div>
        <button
          onClick={() => window.location.reload()}
          className="px-8 py-3 bg-[#d4a359] hover:bg-[#c39248] text-black font-bold rounded-full transition-all transform hover:scale-105"
        >
          Retry Connection
        </button>
      </div>
    );
  }

  return (
    <div className="relative w-full h-screen bg-[#000000] overflow-hidden font-sans text-white">
      {/* Main 2D Graph */}
      <GraphView
        ref={graphViewRef}
        data={filteredData}
        onNodeClick={handleNodeClick}
        selectedNodeId={selectedNode?.id || null}
        forces={forces}
        display={display}
        groups={groups}
      />

      {/* UI Overlay */}
      <div className="absolute inset-0 pointer-events-none z-20">

        {/* Top Navigation Bar */}
        <div className="absolute top-6 left-6 flex items-center gap-4 pointer-events-auto z-50">
          <SearchBar nodes={filteredData.nodes} onSelect={handleNodeClick} />

          <button
            onClick={() => setGraphMode(p => p === 'watchlist' ? 'all' : 'watchlist')}
            className={`flex items-center gap-2 px-4 py-3 rounded-full border transition-all shadow-xl ${graphMode === 'watchlist'
              ? 'bg-[#d4a359] text-black border-[#d4a359]'
              : 'bg-[#1a1a1a]/80 backdrop-blur-md border-white/10 text-white/80 hover:bg-white/10'
              }`}
          >
            <Bookmark size={18} />
            <span className="text-sm font-medium">Watchlist</span>
          </button>

          <button
            onClick={() => setGraphMode(p => p === 'recommendations' ? 'all' : 'recommendations')}
            className={`flex items-center gap-2 px-4 py-3 rounded-full border transition-all shadow-xl ${graphMode === 'recommendations'
              ? 'bg-[#d4a359] text-black border-[#d4a359]'
              : 'bg-[#1a1a1a]/80 backdrop-blur-md border-white/10 text-white/80 hover:bg-white/10'
              }`}
          >
            <Sparkles size={18} />
            <span className="text-sm font-medium">For You</span>
          </button>
        </div>

        {/* Animate Button (replaces branding) */}
        <div className="absolute bottom-8 left-8 pointer-events-auto">
          <motion.button
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5 }}
            onClick={() => graphViewRef.current?.reheat()}
            className="group flex items-center gap-3 px-5 py-3 rounded-full bg-[#1a1a1a]/80 backdrop-blur-md border border-white/10 hover:border-[#d4a359]/50 hover:bg-[#d4a359]/10 transition-all duration-300 shadow-xl"
          >
            <Wand2 size={18} className="text-[#d4a359] group-hover:rotate-12 transition-transform duration-300" />
            <span className="text-sm font-medium text-white/80 group-hover:text-white transition-colors">Animate</span>
          </motion.button>
        </div>

        {/* Controls Panel */}
        <div className="pointer-events-auto">
          <ControlsPanel
            forces={forces} setForces={setForces}
            filters={filters} setFilters={setFilters}
            display={display} setDisplay={setDisplay}
            groups={groups} setGroups={setGroups}
          />
        </div>
      </div>

      {/* Sidebar Panel */}
      <div className="pointer-events-auto z-50 relative">
        <Sidebar
          node={selectedNode}
          onClose={handleCloseSidebar}
          watchlist={watchlist}
          onUpdateWatchlist={handleUpdateWatchlist}
          graphData={graphData}
        />
      </div>
    </div>
  );
}
