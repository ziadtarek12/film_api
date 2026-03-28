import React, { useRef, useEffect, useState, useMemo, useCallback, forwardRef, useImperativeHandle } from 'react';
import ForceGraph2D from 'react-force-graph-2d';
import { GraphData, Node, Link } from '../services/api';
import { FilterGroup } from '../App';

export interface GraphViewHandle {
  reheat: () => void;
}

interface GraphViewProps {
  data: GraphData;
  onNodeClick: (node: Node) => void;
  selectedNodeId: string | null;
  forces: {
    centerForce: number;
    repelForce: number;
    linkForce: number;
    linkDistance: number;
  };
  display: {
    showLabels: boolean;
    nodeSize: 'influence' | 'rating';
    linkThickness: number;
  };
  groups: FilterGroup[];
}

const GraphView = forwardRef<GraphViewHandle, GraphViewProps>(function GraphView(
  { data, onNodeClick, selectedNodeId, forces, display, groups }, ref
) {
  const fgRef = useRef<any>();
  const [dimensions, setDimensions] = useState({ width: window.innerWidth, height: window.innerHeight });
  const [hoverNode, setHoverNode] = useState<Node | null>(null);

  // Incremental animation: null = show all, Set = only show these node IDs
  const [animVisibleIds, setAnimVisibleIds] = useState<Set<string> | null>(null);
  const animTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Expose reheat to parent — incremental reveal animation
  useImperativeHandle(ref, () => ({
    reheat: () => {
      if (!fgRef.current) return;

      // Clear any existing animation
      if (animTimerRef.current) clearInterval(animTimerRef.current);

      // Sort nodes: genres first, then movies, then people
      const sorted = [...data.nodes].sort((a, b) => {
        const order = { Genre: 0, Movie: 1, Person: 2 };
        return (order[a.type] ?? 3) - (order[b.type] ?? 3);
      });

      // Start with empty set
      const revealed = new Set<string>();
      setAnimVisibleIds(new Set(revealed));
      let index = 0;

      // Add nodes one by one every 30ms
      animTimerRef.current = setInterval(() => {
        if (index >= sorted.length) {
          // Done — show all, clear timer
          if (animTimerRef.current) clearInterval(animTimerRef.current);
          animTimerRef.current = null;
          setAnimVisibleIds(null);
          return;
        }

        // Add next node and give it a velocity kick
        const node = sorted[index] as any;
        revealed.add(node.id);
        node.vx = (Math.random() - 0.5) * 40;
        node.vy = (Math.random() - 0.5) * 40;
        index++;

        setAnimVisibleIds(new Set(revealed));

        // Gently reheat every few nodes
        if (index % 5 === 0 && fgRef.current) {
          fgRef.current.d3ReheatSimulation();
        }
      }, 30);
    }
  }));

  // Clean up timer on unmount
  useEffect(() => {
    return () => {
      if (animTimerRef.current) clearInterval(animTimerRef.current);
    };
  }, []);

  // Compute the graph data visible during animation
  const animatedData = useMemo(() => {
    if (animVisibleIds === null) return data; // normal mode — show everything

    const visibleNodes = data.nodes.filter(n => animVisibleIds.has(n.id));
    const visibleLinks = data.links.filter(l => {
      const sId = typeof l.source === 'object' ? (l.source as any).id : l.source;
      const tId = typeof l.target === 'object' ? (l.target as any).id : l.target;
      return animVisibleIds.has(sId) && animVisibleIds.has(tId);
    });
    return { nodes: visibleNodes, links: visibleLinks };
  }, [data, animVisibleIds]);

  // Reheat simulation when animated data node count changes
  const prevNodeCount = useRef(animatedData.nodes.length);
  useEffect(() => {
    if (fgRef.current && prevNodeCount.current !== animatedData.nodes.length) {
      prevNodeCount.current = animatedData.nodes.length;
    }
  }, [animatedData.nodes.length]);

  useEffect(() => {
    const handleResize = () => {
      setDimensions({ width: window.innerWidth, height: window.innerHeight });
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Apply forces
  useEffect(() => {
    if (fgRef.current) {
      fgRef.current.d3Force('charge').strength(-forces.repelForce);
      fgRef.current.d3Force('link').distance(forces.linkDistance);

      const centerForce = fgRef.current.d3Force('center');
      if (centerForce && typeof centerForce.strength === 'function') {
        centerForce.strength(forces.centerForce);
      }

      fgRef.current.d3ReheatSimulation();
    }
  }, [forces]);

  // Focus on selected node
  useEffect(() => {
    if (selectedNodeId && fgRef.current) {
      const node = data.nodes.find(n => n.id === selectedNodeId);
      if (node) {
        fgRef.current.centerAt(node.x, node.y, 1000);
        fgRef.current.zoom(4, 1000);
      }
    }
  }, [selectedNodeId, data.nodes]);

  // Pre-calculate neighbors for fast lookup
  const neighbors = useMemo(() => {
    const map = new Map<string, Set<string>>();
    data.nodes.forEach(n => map.set(n.id, new Set()));
    data.links.forEach(l => {
      const sourceId = typeof l.source === 'object' ? (l.source as any).id : l.source;
      const targetId = typeof l.target === 'object' ? (l.target as any).id : l.target;
      map.get(sourceId)?.add(targetId);
      map.get(targetId)?.add(sourceId);
    });
    return map;
  }, [data]);

  const highlightNodes = useMemo(() => {
    const set = new Set<string>();
    if (hoverNode) {
      set.add(hoverNode.id);
      neighbors.get(hoverNode.id)?.forEach(id => set.add(id));
    }
    if (selectedNodeId) {
      set.add(selectedNodeId);
      neighbors.get(selectedNodeId)?.forEach(id => set.add(id));
    }
    return set;
  }, [hoverNode, selectedNodeId, neighbors]);

  const highlightLinks = useMemo(() => {
    const set = new Set<any>();
    if (hoverNode || selectedNodeId) {
      data.links.forEach(l => {
        const sourceId = typeof l.source === 'object' ? (l.source as any).id : l.source;
        const targetId = typeof l.target === 'object' ? (l.target as any).id : l.target;
        if (
          (hoverNode && (sourceId === hoverNode.id || targetId === hoverNode.id)) ||
          (selectedNodeId && (sourceId === selectedNodeId || targetId === selectedNodeId))
        ) {
          set.add(l);
        }
      });
    }
    return set;
  }, [hoverNode, selectedNodeId, data.links]);

  const getNodeRadius = useCallback((node: Node) => {
    if (display.nodeSize === 'rating' && node.type === 'Movie') {
      return Math.max(1, (node.rating || 5) / 2);
    }
    return Math.sqrt(node.val || 1) * 1.5;
  }, [display.nodeSize]);

  const getNodeColor = useCallback((node: Node) => {
    for (const g of groups) {
      if (!g.query) continue;
      const query = g.query.toLowerCase();
      if (query.startsWith('type:')) {
        if (node.type.toLowerCase() === query.substring(5).trim()) return g.color;
      } else {
        if (node.name.toLowerCase().includes(query)) return g.color;
      }
    }
    return '#9ca3af'; // default gray
  }, [groups]);
  // Check if a node belongs to a hidden group
  const isNodeGroupHidden = useCallback((node: Node) => {
    for (const g of groups) {
      if (!g.hidden || !g.query) continue;
      const query = g.query.toLowerCase();
      if (query.startsWith('type:')) {
        if (node.type.toLowerCase() === query.substring(5).trim()) return true;
      } else {
        if (node.name.toLowerCase().includes(query)) return true;
      }
    }
    return false;
  }, [groups]);

  const paintNode = useCallback((node: any, ctx: CanvasRenderingContext2D, globalScale: number) => {
    const isSelected = selectedNodeId === node.id;
    const isHovered = hoverNode?.id === node.id;
    const isHighlighted = highlightNodes.has(node.id);
    const isDimmedByHover = (hoverNode || selectedNodeId) && !isHighlighted;
    const isHiddenByGroup = isNodeGroupHidden(node);
    const isDimmed = isDimmedByHover || isHiddenByGroup;

    const color = getNodeColor(node);
    const radius = getNodeRadius(node);

    ctx.beginPath();
    ctx.arc(node.x, node.y, radius, 0, 2 * Math.PI, false);

    // Fill opacity logic
    if (isHiddenByGroup) {
      ctx.fillStyle = `${color}4D`; // 30% opacity
    } else if (isDimmedByHover) {
      ctx.fillStyle = `${color}20`; // ~12% opacity
    } else {
      ctx.fillStyle = color;
    }
    ctx.fill();

    // Border for selected/hovered
    if ((isSelected || isHovered) && !isHiddenByGroup) {
      ctx.strokeStyle = '#ffffff';
      ctx.lineWidth = 1 / globalScale;
      ctx.stroke();
    }

    // Text
    const showText = !isHiddenByGroup && (display.showLabels || isHighlighted || globalScale > 2.5);
    if (showText && !isDimmedByHover) {
      const fontSize = isHighlighted ? 12 / globalScale : 8 / globalScale;
      ctx.font = `${fontSize}px Inter, sans-serif`;
      ctx.fillStyle = 'rgba(255,255,255,0.8)';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'top';
      ctx.fillText(node.name, node.x, node.y + radius + 2 / globalScale);
    }
  }, [selectedNodeId, hoverNode, highlightNodes, display.showLabels, getNodeRadius, getNodeColor, isNodeGroupHidden]);

  return (
    <div className="absolute inset-0 bg-[#000000] overflow-hidden">
      <ForceGraph2D
        ref={fgRef}
        graphData={animatedData}
        width={dimensions.width}
        height={dimensions.height}
        nodeLabel={() => ''} // Disable default tooltip
        nodeCanvasObject={paintNode}
        nodePointerAreaPaint={(node: any, color, ctx) => {
          const radius = getNodeRadius(node);
          ctx.fillStyle = color;
          ctx.beginPath();
          ctx.arc(node.x, node.y, radius + 4, 0, 2 * Math.PI, false);
          ctx.fill();
        }}
        linkColor={(link: any) => {
          const sId = typeof link.source === 'object' ? link.source.id : link.source;
          const tId = typeof link.target === 'object' ? link.target.id : link.target;
          const sNode = animatedData.nodes.find(n => n.id === sId);
          const tNode = animatedData.nodes.find(n => n.id === tId);
          if ((sNode && isNodeGroupHidden(sNode)) || (tNode && isNodeGroupHidden(tNode))) return 'rgba(255,255,255,0.03)';
          return highlightLinks.has(link) ? 'rgba(255,255,255,0.6)' : 'rgba(255,255,255,0.1)';
        }}
        linkWidth={(link: any) => {
          const sId = typeof link.source === 'object' ? link.source.id : link.source;
          const tId = typeof link.target === 'object' ? link.target.id : link.target;
          const sNode = animatedData.nodes.find(n => n.id === sId);
          const tNode = animatedData.nodes.find(n => n.id === tId);
          if ((sNode && isNodeGroupHidden(sNode)) || (tNode && isNodeGroupHidden(tNode))) return 0.1;
          return highlightLinks.has(link) ? display.linkThickness * 2 : display.linkThickness;
        }}
        linkDirectionalArrowLength={0}
        onNodeClick={onNodeClick}
        onNodeHover={(node: any) => {
          document.body.style.cursor = node ? 'pointer' : 'default';
          setHoverNode(node || null);
        }}
        backgroundColor="#000000"
        d3VelocityDecay={0.15}
        d3AlphaDecay={0.008}
        warmupTicks={0}
        cooldownTime={15000}
      />
    </div>
  );
});

export default GraphView;
