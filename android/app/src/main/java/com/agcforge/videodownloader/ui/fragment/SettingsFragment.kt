package com.agcforge.videodownloader.ui.fragment

import android.content.Context
import android.content.Intent
import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ArrayAdapter
import android.widget.ImageView
import android.widget.TextView
import androidx.appcompat.app.AlertDialog
import androidx.fragment.app.Fragment
import androidx.lifecycle.lifecycleScope
import com.agcforge.videodownloader.R
import com.agcforge.videodownloader.databinding.FragmentSettingsBinding
import com.agcforge.videodownloader.ui.activities.MainActivity
import com.agcforge.videodownloader.utils.PreferenceManager
import com.agcforge.videodownloader.utils.applyTheme
import com.agcforge.videodownloader.utils.showToast
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import java.util.Locale

class SettingsFragment : Fragment() {

    private var _binding: FragmentSettingsBinding? = null
    private val binding get() = _binding!!

    private lateinit var preferenceManager: PreferenceManager

    private val languages = arrayOf("English", "Indonesia", "Español", "Français", "Português", "中文", "日本語", "العربية", "Deutsch", "हिन्दी", "Русский")
    private val languageCodes = arrayOf("en", "in", "es", "fr", "pt", "zh", "ja", "ar", "de", "hi", "ru")

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
        setupLanguage()
    }

    private fun setupViews() {
        // Load current settings
    }

    private fun setupListeners() {
        binding.apply {
            btnQualitySettings.setOnClickListener { showQualitySettings() }
            btnStorageSettings.setOnClickListener { showStorageSettings() }
            btnClearCache.setOnClickListener { clearCache() }
            btnAbout.setOnClickListener { showAboutDialog() }
            btnLanguage.setOnClickListener { showLanguageDialog() }
        }
    }

    private fun showQualitySettings() {
        val qualities = arrayOf("Auto", "1080p", "720p", "480p", "360p")

        AlertDialog.Builder(requireContext())
            .setTitle(getString(R.string.quality_dialog_title))
            .setItems(qualities) { dialog, which ->
                requireContext().showToast(getString(R.string.quality_set_toast, qualities[which]))
                dialog.dismiss()
            }
            .show()
    }

    private fun showStorageSettings() {
        requireContext().showToast(getString(R.string.storage_settings_soon))
    }

    private fun clearCache() {
        try {
            requireContext().cacheDir.deleteRecursively()
            requireContext().showToast(getString(R.string.cache_cleared_success))
        } catch (e: Exception) {
            requireContext().showToast(getString(R.string.cache_cleared_fail))
        }
    }

    private fun showAboutDialog() {
        AlertDialog.Builder(requireContext())
            .setTitle(getString(R.string.about))
            .setMessage(getString(R.string.about_dialog_message))
            .setPositiveButton(getString(R.string.ok), null)
            .show()
    }

    private fun setupThemeSwitch() {
        lifecycleScope.launch {
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

    private fun setupLanguage() {
        lifecycleScope.launch {
            val currentLangCode = preferenceManager.language.first() ?: Locale.getDefault().language
            val currentLangIndex = languageCodes.indexOf(currentLangCode)
            if (currentLangIndex != -1) {
                binding.tvCurrentLanguage.text = languages[currentLangIndex]
            }
        }
    }

    private fun showLanguageDialog() {
        val currentLangCode = runBlocking { preferenceManager.language.first() } ?: Locale.getDefault().language
        val checkedItem = languageCodes.indexOf(currentLangCode)

        val adapter = LanguageAdapter(requireContext(), languages, languageCodes)

        AlertDialog.Builder(requireContext())
            .setTitle(getString(R.string.language))
            .setSingleChoiceItems(adapter, checkedItem) { dialog, which ->
                val selectedLangCode = languageCodes[which]
                lifecycleScope.launch {
                    preferenceManager.saveLanguage(selectedLangCode)
                    dialog.dismiss()

                    // Restart the app to apply the new language
                    val intent = Intent(requireActivity(), MainActivity::class.java)
                    intent.addFlags(Intent.FLAG_ACTIVITY_CLEAR_TOP or Intent.FLAG_ACTIVITY_NEW_TASK)
                    startActivity(intent)
                    requireActivity().finish()
                }
            }
            .setNegativeButton(getString(R.string.cancel), null)
            .show()
    }

    private inner class LanguageAdapter(
        context: Context,
        private val languages: Array<String>,
        private val languageCodes: Array<String>
    ) : ArrayAdapter<String>(context, R.layout.list_item_language, R.id.tvLanguageName, languages) {

        override fun getView(position: Int, convertView: View?, parent: ViewGroup): View {
            val view = super.getView(position, convertView, parent)
            val ivFlag = view.findViewById<ImageView>(R.id.ivFlag)
            val tvLanguageName = view.findViewById<TextView>(R.id.tvLanguageName)

            tvLanguageName.text = languages[position]

            val flagResId = context.resources.getIdentifier("flag_${languageCodes[position]}", "drawable", context.packageName)
            if (flagResId != 0) {
                ivFlag.setImageResource(flagResId)
            } else {
                ivFlag.setImageResource(R.drawable.ic_language) // Default icon
            }

            return view
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
