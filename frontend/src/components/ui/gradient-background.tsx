'use client';
import type React from 'react';
import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';

type GradientBackgroundProps = React.ComponentProps<'div'> & {
	gradients?: string[];
	animationDuration?: number;
	animationDelay?: number;
	enableCenterContent?: boolean;
	overlay?: boolean;
	overlayOpacity?: number;
};

const Default_Gradients = [
	"linear-gradient(135deg, #0D1B3E 0%, #0E7490 100%)",
	"linear-gradient(135deg, #1A0A40 0%, #059669 100%)",
	"linear-gradient(135deg, #0D1B3E 0%, #6D28D9 100%)",
	"linear-gradient(135deg, #2D0A20 0%, #BE123C 100%)",
	"linear-gradient(135deg, #0D1B3E 0%, #0E7490 100%)",
];

export function GradientBackground({
	children,
	className = '',
	gradients = Default_Gradients,
	animationDuration = 8,
	animationDelay = 0.5,
	overlay = false,
	overlayOpacity = 0.3,
}: GradientBackgroundProps) {
	return (
		<div className={cn('w-full relative min-h-screen overflow-hidden', className)}>
			<motion.div
				className="absolute inset-0"
				style={{ background: gradients[0] }}
				animate={{ background: gradients }}
				transition={{
					delay: animationDelay,
					duration: animationDuration,
					repeat: Number.POSITIVE_INFINITY,
					ease: 'easeInOut',
				}}
			/>

			{overlay && (
				<div
					className="absolute inset-0 bg-black"
					style={{ opacity: overlayOpacity }}
				/>
			)}

			{children && (
				<div className="relative z-10 flex min-h-screen items-center justify-center">
					{children}
				</div>
			)}
		</div>
	);
}
