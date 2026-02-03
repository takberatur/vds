package com.agcforge.videodownloader.ui.fragment

import android.Manifest
import android.app.DownloadManager
import android.content.Context
import android.content.pm.PackageManager
import android.annotation.SuppressLint
import android.content.ClipboardManager
import android.net.Uri
import android.os.Bundle
import android.os.Environment
import android.text.Editable
import android.text.TextWatcher
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Toast
import androidx.activity.result.contract.ActivityResultContracts
import androidx.core.content.ContextCompat
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.lifecycle.lifecycleScope
import androidx.navigation.fragment.findNavController
import androidx.recyclerview.widget.GridLayoutManager
import com.agcforge.videodownloader.R
import com.agcforge.videodownloader.data.model.DownloadFormat
import com.agcforge.videodownloader.data.model.DownloadTask
import com.agcforge.videodownloader.data.model.Platform
import com.agcforge.videodownloader.databinding.FragmentHomeBinding
import com.agcforge.videodownloader.ui.adapter.PlatformAdapter
import com.agcforge.videodownloader.ui.component.FormatSelectionDialog
import com.agcforge.videodownloader.ui.viewmodel.HomeViewModel
import com.agcforge.videodownloader.utils.AppManager
import com.agcforge.videodownloader.utils.DownloadManagerCleaner
import com.agcforge.videodownloader.utils.PreferenceManager
import com.agcforge.videodownloader.utils.Resource
import com.agcforge.videodownloader.utils.showToast
import com.google.android.material.dialog.MaterialAlertDialogBuilder
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import androidx.core.net.toUri

class HomeFragment : Fragment() {

    private var _binding: FragmentHomeBinding? = null
    private val binding get() = _binding!!

    private val viewModel: HomeViewModel by viewModels()
    private lateinit var platformAdapter: PlatformAdapter
    private var selectedPlatform: Platform? = null
	private var isSubmitting = false
	private lateinit var preferenceManager: PreferenceManager
	private var pendingDownloadUrl: String? = null

	private val storagePermissionLauncher = registerForActivityResult(
		ActivityResultContracts.RequestPermission()
	) { granted ->
		val url = pendingDownloadUrl
		pendingDownloadUrl = null
		if (granted && url != null) {
			enqueueDownload(url)
		} else if (!granted) {
			requireContext().showToast(getString(R.string.storage_permission_denied))
		}
	}

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = FragmentHomeBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
		preferenceManager = PreferenceManager(requireContext())

        setupRecyclerView()
        setupListeners()
        observeViewModel()
        observeHistoryNavigation()

        viewModel.loadPlatforms()
		updateDownloadButtonState()
    }

    private fun observeHistoryNavigation() {
        findNavController().currentBackStackEntry?.savedStateHandle?.getLiveData<String>("history_url")
            ?.observe(viewLifecycleOwner) { url ->
                binding.etUrl.setText(url)
                findNavController().currentBackStackEntry?.savedStateHandle?.remove<String>("history_url")
            }
    }

    @SuppressLint("SetTextI18n")
    private fun setupRecyclerView() {
        platformAdapter = PlatformAdapter { platform ->
            // Handle platform click - auto-fill platform info
            selectedPlatform = platform
            platformAdapter.setSelection(platform)
            binding.tvPlatformName.text = "${getString(R.string.download_video)} - ${platform.name}"
            requireContext().showToast("Selected: ${platform.name}")
			updateDownloadButtonState()
        }

        binding.rvPlatforms.apply {
            adapter = platformAdapter
            layoutManager = GridLayoutManager(requireContext(), 2)
        }
    }

    private fun setupListeners() {
        binding.btnDownload.setOnClickListener {
            val url = binding.etUrl.text.toString().trim()
			if (!isUrlValid(url)) {
				binding.tilUrl.error = getString(R.string.please_enter_valid_url)
				updateDownloadButtonState()
				return@setOnClickListener
			}

            val platform = selectedPlatform
            if (platform == null) {
                requireContext().showToast(getString(R.string.please_select_a_platform))
                return@setOnClickListener
            }

			binding.tilUrl.error = null
            viewModel.createDownload(url, platform.type)
            addToHistory(url)
        }

		binding.etUrl.addTextChangedListener(object : TextWatcher {
			override fun beforeTextChanged(s: CharSequence?, start: Int, count: Int, after: Int) = Unit
			override fun onTextChanged(s: CharSequence?, start: Int, before: Int, count: Int) = Unit
			override fun afterTextChanged(s: Editable?) {
				if (binding.tilUrl.error != null) binding.tilUrl.error = null
				updateDownloadButtonState()
			}
		})

        binding.etUrl.setOnFocusChangeListener { _, hasFocus ->
            if (!hasFocus) {
                binding.tilUrl.error = null
            }
        }

        binding.btnPaste.setOnClickListener {
            val clipboard = requireContext().getSystemService(ClipboardManager::class.java)
            val clip = clipboard.primaryClip
            if (clip != null && clip.itemCount > 0) {
                val text = clip.getItemAt(0).text?.toString()
                if (!text.isNullOrEmpty()) {
                    binding.etUrl.setText(text)
                    requireContext().showToast("Pasted: $text", Toast.LENGTH_SHORT)
					updateDownloadButtonState()
                }
            }
        }
    }

    private fun observeViewModel() {
        viewLifecycleOwner.lifecycleScope.launch {
            viewModel.platforms.collect { resource ->
                when (resource) {
                    is Resource.Loading -> {
                        binding.progressBar.visibility = View.VISIBLE
                    }
                    is Resource.Success -> {
                        binding.progressBar.visibility = View.GONE
                        resource.data?.let { platformAdapter.submitList(it) }
                    }
                    is Resource.Error -> {
                        binding.progressBar.visibility = View.GONE
                        requireContext().showToast(resource.message ?: getString(R.string.failed_to_load_platform))
                    }

                    else -> {}
                }
            }
        }

        viewLifecycleOwner.lifecycleScope.launch {
            viewModel.downloadResult.collect { resource ->
                when (resource) {
                    is Resource.Loading -> {
						isSubmitting = true
                        binding.btnDownload.isEnabled = false
                        binding.progressBar.visibility = View.VISIBLE
                    }
                    is Resource.Success -> {
						isSubmitting = false
                        binding.btnDownload.isEnabled = true
                        binding.progressBar.visibility = View.GONE

                        resource.data?.let { task ->
                            // Show format selection dialog if formats available
                            if (!task.formats.isNullOrEmpty()) {
                                showFormatSelectionDialog(task)
                            } else {
							val directUrl = task.filePath?.let { sanitizeApiUrl(it) }
							if (!directUrl.isNullOrBlank() && isUrlValid(directUrl)) {
								enqueueDownload(directUrl)
								requireContext().showToast(getString(R.string.download_started))
								binding.etUrl.text?.clear()
								updateDownloadButtonState()
							} else {
								requireContext().showToast(getString(R.string.no_format_available_to_download))
								updateDownloadButtonState()
							}
                            }
                        }
                    }
                    is Resource.Error -> {
						isSubmitting = false
                        binding.btnDownload.isEnabled = true
                        binding.progressBar.visibility = View.GONE
                        requireContext().showToast(resource.message ?: getString(R.string.download_failed))
						updateDownloadButtonState()
                    }
					is Resource.Idle -> {
						isSubmitting = false
						binding.progressBar.visibility = View.GONE
						updateDownloadButtonState()
					}
                }
            }
        }
    }

	private fun isUrlValid(url: String): Boolean {
		if (url.isBlank()) return false
		if (!url.startsWith("http://") && !url.startsWith("https://")) return false
		return runCatching { url.toUri() }.isSuccess
	}

	private fun sanitizeApiUrl(raw: String): String {
		return raw.trim()
			.trim('`')
			.trim('"')
			.trim('\'')
			.trim()
	}

	private fun updateDownloadButtonState() {
		if (isSubmitting) {
			binding.btnDownload.isEnabled = false
			return
		}

		val url = binding.etUrl.text?.toString()?.trim().orEmpty()
		val enabled = url.isNotBlank()
		binding.btnDownload.isEnabled = enabled
	}

    private fun showFormatSelectionDialog(task: DownloadTask) {
        val formats = task.formats
        if (formats.isNullOrEmpty()) {
            requireContext().showToast(getString(R.string.no_format_available_to_download))
            return
        }

        FormatSelectionDialog.Builder(requireContext())
            .setTask(task)
            .setOnFormatSelected { selectedFormat ->
                // Handle format selected
                enqueueDownload(buildProxyVideoUrl(task, selectedFormat))
                requireContext().showToast(getString(R.string.download_started))
                binding.etUrl.text?.clear()
                updateDownloadButtonState()
            }
            .show()
    }

	private fun buildProxyVideoUrl(
		task: DownloadTask,
		format: DownloadFormat
	): String {
		val base = AppManager.baseUrl
		val endpoint = if (base.endsWith("/")) {
			"${base}public-proxy/downloads/file/video"
		} else {
			"$base/public-proxy/downloads/file/video"
		}

		val formatId = format.formatId
		val resolution = format.height?.let { "${it}p" }
		val effectiveFormat = formatId ?: resolution

		val filename = task.title?.takeIf { it.isNotBlank() } ?: getString(R.string.download)
		val uriBuilder = endpoint.toUri().buildUpon()
			.appendQueryParameter("task_id", task.id)
			.appendQueryParameter("filename", filename)
		if (!effectiveFormat.isNullOrBlank()) {
			uriBuilder.appendQueryParameter("format_id", effectiveFormat)
		}
		return uriBuilder.build().toString()
	}

	private fun enqueueDownload(url: String) {
		viewLifecycleOwner.lifecycleScope.launch {
			DownloadManagerCleaner.clearFailedDownloads(requireContext())
			val storageLocation = preferenceManager.storageLocation.first() ?: "app"
			val uri = url.toUri()
			val fileName = deriveDownloadFileName(uri)

			if (storageLocation == "downloads" && android.os.Build.VERSION.SDK_INT < 29) {
				val granted = ContextCompat.checkSelfPermission(
					requireContext(),
					Manifest.permission.WRITE_EXTERNAL_STORAGE
				) == PackageManager.PERMISSION_GRANTED
				if (!granted) {
					pendingDownloadUrl = url
					storagePermissionLauncher.launch(Manifest.permission.WRITE_EXTERNAL_STORAGE)
					return@launch
				}
			}

			val dm = requireContext().getSystemService(Context.DOWNLOAD_SERVICE) as DownloadManager
			val request = DownloadManager.Request(uri)
				.setTitle(fileName)
				.setNotificationVisibility(DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED)
				.setAllowedOverMetered(true)
				.setAllowedOverRoaming(true)

			if (storageLocation == "downloads") {
				request.setDestinationInExternalPublicDir(Environment.DIRECTORY_DOWNLOADS, fileName)
			} else {
				request.setDestinationInExternalFilesDir(requireContext(), Environment.DIRECTORY_DOWNLOADS, fileName)
			}

			try {
				dm.enqueue(request)
				requireContext().showToast(getString(R.string.download_started))
			} catch (e: Exception) {
				requireContext().showToast(e.message ?: getString(R.string.download_failed))
			}
		}
	}

	private fun deriveDownloadFileName(uri: android.net.Uri): String {
		val defaultBase = "video_${System.currentTimeMillis()}"
		val filenameParam = uri.getQueryParameter("filename")?.takeIf { it.isNotBlank() }
		val lastSegment = uri.lastPathSegment?.takeIf { it.isNotBlank() }
		val baseNameRaw = filenameParam ?: lastSegment ?: defaultBase
		val taskIdSuffix = uri.getQueryParameter("task_id")?.takeIf { it.isNotBlank() }?.take(8)

		val extFromBase = baseNameRaw.substringAfterLast('.', missingDelimiterValue = "").lowercase()
		val expectedExt = when {
			uri.path?.endsWith("/mp3", ignoreCase = true) == true -> "mp3"
			extFromBase == "mp3" || extFromBase == "mp4" -> extFromBase
			else -> "mp4"
		}

		val baseWithoutExt = if (baseNameRaw.contains('.')) baseNameRaw.substringBeforeLast('.') else baseNameRaw
		val safeBase = baseWithoutExt
			.replace(Regex("[\\\\/:*?\"<>|]"), "_")
			.replace(Regex("\n|\r|\t"), " ")
			.trim()
			.ifBlank { defaultBase }
			.let { base ->
				if (!taskIdSuffix.isNullOrBlank()) {
					"${base}_$taskIdSuffix"
				} else {
					base
				}
			}
			.take(80)

		return "$safeBase.$expectedExt"
	}

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }

    private fun handleFormatSelection(task: DownloadTask, selectedFormat: DownloadFormat) {
        val downloadUrl = buildProxyVideoUrl(task, selectedFormat)

        MaterialAlertDialogBuilder(requireContext())
            .setTitle(getString(R.string.confirm_download))
            .setMessage(getString(R.string.confirm_download_message, selectedFormat.getQualityLabel(), selectedFormat.getFormatDescription()))
            .setPositiveButton(getString(R.string.yes)) { _, _ ->
                enqueueDownload(downloadUrl)
                requireContext().showToast(getString(R.string.download_started))
                binding.etUrl.text?.clear()
                updateDownloadButtonState()
            }
            .setNegativeButton(getString(R.string.no), null)
            .show()
    }

    private fun addToHistory(url: String) {
        viewLifecycleOwner.lifecycleScope.launch {
            preferenceManager.addToHistory(url)
        }
    }
}
