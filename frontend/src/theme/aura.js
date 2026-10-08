import Aura from '@primeuix/themes/aura'
import { definePreset } from '@primeuix/themes'

export const WarmAura = definePreset(Aura, {
  semantic: {
    primary: {
      50: '#FDF6EC',
      100: '#F5E6CC',
      200: '#E8C9A0',
      300: '#D4A76A',
      400: '#B07D3A',
      500: '#8B5A2B',
      600: '#6F4620',
      700: '#5A381A',
      800: '#4A2E15',
      900: '#3D2612',
      950: '#25160A',
    },
    colorScheme: {
      light: {
        primary: {
          color: '{primary.500}',
          inverseColor: '#ffffff',
          hoverColor: '{primary.600}',
          activeColor: '{primary.700}',
        },
        surface: {
          0: '#ffffff',
          50: '#F9FAFB',
          100: '#F3F4F6',
          200: '#E5E7EB',
          300: '#D1D5DB',
          400: '#9CA3AF',
          500: '#6B7280',
          600: '#4B5563',
          700: '#374151',
          800: '#1F2937',
          900: '#111827',
          950: '#030712',
        },
      },
    },
  },
  components: {
    card: {
      colorScheme: {
        light: {
          root: { background: '#ffffff', color: '{surface.800}' },
        },
      },
    },
  },
})
