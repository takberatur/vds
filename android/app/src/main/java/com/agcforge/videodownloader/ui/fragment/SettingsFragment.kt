package com.agcforge.videodownloader.ui.fragment

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.Fragment
import androidx.lifecycle.lifecycleScope
import com.agcforge.videodownloader.databinding.FragmentSettingsBinding
import com.agcforge.videodownloader.utils.PreferenceManager
import com.agcforge.videodownloader.utils.applyTheme
import com.agcforge.videodownloader.utils.showToast
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

class SettingsFragment : Fragment() {

    private var _binding: FragmentSettingsBinding? = null
    private val binding get() = _binding!!

    private lateinit var preferenceManager: PreferenceManager

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = FragmentSettingsBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        preferenceManager = PreferenceManager(requireContext())

        setupViews()
        setupListeners()
        setupThemeSwitch()
    }

    private fun setupViews() {
        // Load current settings
        binding.apply {
            // Set initial values from preferences
            // Example: switchAutoDownload.isChecked = preferenceManager.getAutoDownload()
        }
    }

    private fun setupListeners() {
        binding.apply {
            // Quality settings
            btnQualitySettings.setOnClickListener {
                showQualitySettings()
            }

            // Storage settings
            btnStorageSettings.setOnClickListener {
                showStorageSettings()
            }

            // Clear cache
            btnClearCache.setOnClickListener {
                clearCache()
            }

            // About
            btnAbout.setOnClickListener {
                showAboutDialog()
            }
        }
    }

    private fun showQualitySettings() {
        val qualities = arrayOf("Auto", "1080p", "720p", "480p", "360p")

        androidx.appcompat.app.AlertDialog.Builder(requireContext())
            .setTitle("Default Download Quality")
            .setItems(qualities) { dialog, which ->
                requireContext().showToast("Quality set to ${qualities[which]}")
                dialog.dismiss()
            }
            .show()
    }

    private fun showStorageSettings() {
        requireContext().showToast("Storage settings coming soon")
    }

    private fun clearCache() {
        try {
            requireContext().cacheDir.deleteRecursively()
            requireContext().showToast("Cache cleared successfully")
        } catch (e: Exception) {
            requireContext().showToast("Failed to clear cache")
        }
    }

    private fun showAboutDialog() {
        androidx.appcompat.app.AlertDialog.Builder(requireContext())
            .setTitle("About")
            .setMessage(
                "Video Downloader\n" +
                        "Version 1.0\n\n" +
                        "Download videos from various platforms including YouTube, Instagram, TikTok, and more.\n\n" +
                        "© 2026 AGCForge"
            )
            .setPositiveButton("OK", null)
            .show()
    }

    private fun setupThemeSwitch() {
        lifecycleScope.launch {
            // Set the switch to the current theme
            val currentTheme = preferenceManager.theme.first()
            binding.switchTheme.isChecked = currentTheme == "Dark"
        }

        binding.switchTheme.setOnCheckedChangeListener { _, isChecked ->
            val newTheme = if (isChecked) "Dark" else "Light"
            lifecycleScope.launch {
                preferenceManager.saveTheme(newTheme)
                applyTheme(newTheme)
            }
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
