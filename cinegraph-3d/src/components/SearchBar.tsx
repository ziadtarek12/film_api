import React, { useState } from 'react';
import { Search } from 'lucide-react';
import { Node } from '../services/api';

interface SearchBarProps {
  nodes: Node[];
  onSelect: (node: Node) => void;
}

export default function SearchBar({ nodes, onSelect }: SearchBarProps) {
  const [query, setQuery] = useState('');
  const [isOpen, setIsOpen] = useState(false);

  const filteredNodes = query
    ? nodes.filter(n => n.name.toLowerCase().includes(query.toLowerCase()))
    : [];

  return (
    <div className="relative w-80 z-50">
      <div className="relative">
        <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
          <Search size={18} className="text-white/50" />
        </div>
        <input
          type="text"
          className="w-full bg-[#1a1a1a]/80 backdrop-blur-md border border-white/10 rounded-full py-3 pl-12 pr-4 text-white placeholder-white/30 focus:outline-none focus:ring-2 focus:ring-white/20 transition-all font-mono text-sm shadow-xl"
          placeholder="Search films, actors, directors..."
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setIsOpen(true);
          }}
          onFocus={() => setIsOpen(true)}
          onBlur={() => setTimeout(() => setIsOpen(false), 200)}
        />
      </div>

      {isOpen && filteredNodes.length > 0 && (
        <div className="absolute mt-2 w-full bg-[#1a1a1a]/95 backdrop-blur-xl border border-white/10 rounded-2xl overflow-hidden shadow-2xl">
          <ul className="max-h-60 overflow-y-auto py-2">
            {filteredNodes.map(node => (
              <li key={node.id}>
                <button
                  className="w-full text-left px-4 py-3 hover:bg-white/5 transition-colors flex items-center justify-between group"
                  onClick={() => {
                    onSelect(node);
                    setQuery('');
                    setIsOpen(false);
                  }}
                >
                  <span className="text-white/80 group-hover:text-white transition-colors">{node.name}</span>
                  <span className="text-xs font-mono text-white/30 uppercase tracking-widest">
                    {node.type}
                  </span>
                </button>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
