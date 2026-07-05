export interface LogoProps {
  size?: number;
}

export function Logo({ size = 34 }: LogoProps) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 200 200"
      width={size}
      height={size}
      aria-hidden="true"
      focusable="false"
    >
      <defs>
        <filter id="cc-neon-glow" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="5" result="coloredBlur" />
          <feMerge>
            <feMergeNode in="coloredBlur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>
      <polygon
        points="100,15 175,55 175,145 100,185 25,145 25,55"
        fill="#06130c"
        stroke="#34d399"
        strokeWidth="6"
        filter="url(#cc-neon-glow)"
      />
      <path
        d="M 132,68 A 45 45 0 1 0 132,132"
        fill="none"
        stroke="#34d399"
        strokeWidth="12"
        strokeLinecap="round"
      />
      <path
        d="M 140,84 L 144,96 L 156,100 L 144,104 L 140,116 L 136,104 L 124,100 L 136,96 Z"
        fill="#a7f3d0"
      />
    </svg>
  );
}
