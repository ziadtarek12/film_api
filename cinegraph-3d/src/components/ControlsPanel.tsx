import React, { useState } from 'react';
import { ChevronDown, ChevronRight, RotateCcw, X, Plus, Eye, EyeOff } from 'lucide-react';
import { FilterGroup } from '../App';

interface ControlsPanelProps {
  forces: {
    centerForce: number;
    repelForce: number;
    linkForce: number;
    linkDistance: number;
  };
  setForces: React.Dispatch<React.SetStateAction<any>>;
  filters: {
    title: string;
    genres: string;
    actors: string;
    directors: string;
  };
  setFilters: React.Dispatch<React.SetStateAction<any>>;
  display: {
    showLabels: boolean;
    nodeSize: 'influence' | 'rating';
    linkThickness: number;
  };
  setDisplay: React.Dispatch<React.SetStateAction<any>>;
  groups: FilterGroup[];
  setGroups: React.Dispatch<React.SetStateAction<FilterGroup[]>>;
}

export default function ControlsPanel({ forces, setForces, filters, setFilters, display, setDisplay, groups, setGroups }: ControlsPanelProps) {
  const [openSections, setOpenSections] = useState({
    filters: true,
    groups: true,
    display: true,
    forces: true,
  });

  const toggleSection = (section: keyof typeof openSections) => {
    setOpenSections(prev => ({ ...prev, [section]: !prev[section] }));
  };

  const resetFilters = (e: React.MouseEvent) => {
    e.stopPropagation();
    setFilters({ title: '', genres: '', actors: '', directors: '' });
  };

  const addGroup = () => {
    const colors = ['#ef4444', '#f97316', '#f59e0b', '#84cc16', '#22c55e', '#06b6d4', '#3b82f6', '#8b5cf6', '#d946ef', '#f43f5e'];
    const randomColor = colors[Math.floor(Math.random() * colors.length)];
    setGroups([...groups, { id: Date.now().toString(), query: '', color: randomColor, hidden: false }]);
  };

  const updateGroup = (id: string, field: keyof FilterGroup, value: any) => {
    setGroups(groups.map(g => g.id === id ? { ...g, [field]: value } : g));
  };

  const removeGroup = (id: string) => {
    setGroups(groups.filter(g => g.id !== id));
  };

  return (
    <div className="absolute right-4 top-4 w-72 bg-[#1e1e1e] border border-white/10 rounded-lg text-white/90 text-sm font-sans flex flex-col max-h-[calc(100vh-2rem)] overflow-y-auto overflow-x-visible shadow-2xl z-40">
      {/* Filters Header */}
      <div className="flex items-center justify-between p-3 border-b border-white/10 hover:bg-white/5 cursor-pointer" onClick={() => toggleSection('filters')}>
        <div className="flex items-center gap-2">
          {openSections.filters ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
          <span className="font-semibold">Filters</span>
        </div>
        <div className="flex items-center gap-2 text-white/50">
          <RotateCcw size={14} className="hover:text-white transition-colors" onClick={resetFilters} />
          <X size={16} className="hover:text-white transition-colors" />
        </div>
      </div>

      {openSections.filters && (
        <div className="px-4 pb-3 pt-2 space-y-3">
          <div>
            <label className="text-xs text-white/50 mb-1 block">Title</label>
            <input
              type="text"
              placeholder="e.g. Inception"
              value={filters.title}
              onChange={(e) => setFilters((prev: any) => ({ ...prev, title: e.target.value }))}
              className="w-full bg-[#2a2a2a] border border-white/10 rounded px-2 py-1.5 text-xs focus:outline-none focus:border-white/30"
            />
          </div>
          <div>
            <label className="text-xs text-white/50 mb-1 block">Genres</label>
            <input
              type="text"
              placeholder="e.g. Sci-Fi, Action"
              value={filters.genres}
              onChange={(e) => setFilters((prev: any) => ({ ...prev, genres: e.target.value }))}
              className="w-full bg-[#2a2a2a] border border-white/10 rounded px-2 py-1.5 text-xs focus:outline-none focus:border-white/30"
            />
          </div>
          <div>
            <label className="text-xs text-white/50 mb-1 block">Actors</label>
            <input
              type="text"
              placeholder="e.g. Leonardo DiCaprio"
              value={filters.actors}
              onChange={(e) => setFilters((prev: any) => ({ ...prev, actors: e.target.value }))}
              className="w-full bg-[#2a2a2a] border border-white/10 rounded px-2 py-1.5 text-xs focus:outline-none focus:border-white/30"
            />
          </div>
          <div>
            <label className="text-xs text-white/50 mb-1 block">Directors</label>
            <input
              type="text"
              placeholder="e.g. Christopher Nolan"
              value={filters.directors}
              onChange={(e) => setFilters((prev: any) => ({ ...prev, directors: e.target.value }))}
              className="w-full bg-[#2a2a2a] border border-white/10 rounded px-2 py-1.5 text-xs focus:outline-none focus:border-white/30"
            />
          </div>
        </div>
      )}

      {/* Groups */}
      <div className="flex items-center gap-2 p-3 border-t border-white/10 hover:bg-white/5 cursor-pointer" onClick={() => toggleSection('groups')}>
        {openSections.groups ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
        <span className="font-semibold">Groups</span>
      </div>
      {openSections.groups && (
        <div className="px-4 pb-3 space-y-2 relative">
          {groups.map(g => (
            <div key={g.id} className={`flex items-center gap-2 transition-opacity duration-200 ${g.hidden ? 'opacity-40' : 'opacity-100'}`}>
              <button
                onClick={() => updateGroup(g.id, 'hidden', !g.hidden)}
                className={`shrink-0 transition-colors ${g.hidden ? 'text-white/30 hover:text-white/60' : 'text-white/50 hover:text-white'}`}
                title={g.hidden ? "Show group" : "Hide group"}
              >
                {g.hidden ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
              <input
                type="text"
                value={g.query}
                onChange={(e) => updateGroup(g.id, 'query', e.target.value)}
                placeholder="e.g. type:Movie or Inception"
                className="flex-1 bg-[#2a2a2a] border border-white/10 rounded px-2 py-1 text-xs focus:outline-none focus:border-white/30 min-w-0"
              />
              <div className="relative shrink-0">
                <input
                  type="color"
                  value={g.color}
                  onChange={(e) => updateGroup(g.id, 'color', e.target.value)}
                  className="w-6 h-6 p-0 border-0 rounded cursor-pointer appearance-none bg-transparent"
                  style={{ backgroundColor: g.color }}
                  title="Change color"
                />
              </div>
              <X size={16} className="text-white/50 hover:text-white cursor-pointer shrink-0" onClick={() => removeGroup(g.id)} title="Remove group" />
            </div>
          ))}
          <button
            onClick={addGroup}
            className="w-full mt-2 py-1.5 bg-[#d4a359] hover:bg-[#e5b46a] text-black font-medium rounded text-xs transition-colors flex items-center justify-center gap-1"
          >
            <Plus size={14} /> New group
          </button>
        </div>
      )}

      {/* Display */}
      <div className="flex items-center gap-2 p-3 border-t border-white/10 hover:bg-white/5 cursor-pointer" onClick={() => toggleSection('display')}>
        {openSections.display ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
        <span className="font-semibold">Display</span>
      </div>
      {openSections.display && (
        <div className="px-4 pb-4 space-y-4">
          <div className="flex items-center justify-between">
            <span className="text-xs">Show Labels</span>
            <input
              type="checkbox"
              checked={display.showLabels}
              onChange={(e) => setDisplay((prev: any) => ({ ...prev, showLabels: e.target.checked }))}
              className="accent-[#d4a359]"
            />
          </div>
          <div>
            <div className="flex justify-between text-xs mb-1">
              <span>Node Size By</span>
            </div>
            <select
              value={display.nodeSize}
              onChange={(e) => setDisplay((prev: any) => ({ ...prev, nodeSize: e.target.value }))}
              className="w-full bg-[#2a2a2a] border border-white/10 rounded px-2 py-1.5 text-xs focus:outline-none focus:border-white/30"
            >
              <option value="influence">Influence (Links)</option>
              <option value="rating">Rating</option>
            </select>
          </div>
          <div>
            <div className="flex justify-between text-xs mb-1">
              <span>Link Thickness</span>
            </div>
            <input
              type="range" min="0.1" max="5" step="0.1"
              value={display.linkThickness}
              onChange={(e) => setDisplay((prev: any) => ({ ...prev, linkThickness: parseFloat(e.target.value) }))}
              className="w-full h-1 bg-white/20 rounded-lg appearance-none cursor-pointer accent-white"
            />
          </div>
        </div>
      )}

      {/* Forces */}
      <div className="flex items-center gap-2 p-3 border-t border-white/10 hover:bg-white/5 cursor-pointer" onClick={() => toggleSection('forces')}>
        {openSections.forces ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
        <span className="font-semibold">Forces</span>
      </div>
      {openSections.forces && (
        <div className="px-4 pb-4 space-y-4">
          <div>
            <div className="flex justify-between text-xs mb-1">
              <span>Center force</span>
            </div>
            <input type="range" min="0" max="1" step="0.01" value={forces.centerForce} onChange={(e) => setForces((prev: any) => ({ ...prev, centerForce: parseFloat(e.target.value) }))} className="w-full h-1 bg-white/20 rounded-lg appearance-none cursor-pointer accent-white" />
          </div>
          <div>
            <div className="flex justify-between text-xs mb-1">
              <span>Repel force</span>
            </div>
            <input type="range" min="0" max="1000" step="10" value={forces.repelForce} onChange={(e) => setForces((prev: any) => ({ ...prev, repelForce: parseFloat(e.target.value) }))} className="w-full h-1 bg-white/20 rounded-lg appearance-none cursor-pointer accent-white" />
          </div>
          <div>
            <div className="flex justify-between text-xs mb-1">
              <span>Link force</span>
            </div>
            <input type="range" min="0" max="1" step="0.01" value={forces.linkForce} onChange={(e) => setForces((prev: any) => ({ ...prev, linkForce: parseFloat(e.target.value) }))} className="w-full h-1 bg-white/20 rounded-lg appearance-none cursor-pointer accent-white" />
          </div>
          <div>
            <div className="flex justify-between text-xs mb-1">
              <span>Link distance</span>
            </div>
            <input type="range" min="1" max="200" step="1" value={forces.linkDistance} onChange={(e) => setForces((prev: any) => ({ ...prev, linkDistance: parseFloat(e.target.value) }))} className="w-full h-1 bg-white/20 rounded-lg appearance-none cursor-pointer accent-white" />
          </div>
        </div>
      )}
    </div>
  );
}
