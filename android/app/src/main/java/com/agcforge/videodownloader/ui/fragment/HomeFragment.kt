package com.agcforge.videodownloader.ui.fragment

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.lifecycle.lifecycleScope
import androidx.recyclerview.widget.GridLayoutManager
import com.agcforge.videodownloader.databinding.FragmentHomeBinding
import com.agcforge.videodownloader.ui.adapter.PlatformAdapter
import com.agcforge.videodownloader.ui.viewmodel.HomeViewModel
import com.agcforge.videodownloader.utils.Resource
import com.agcforge.videodownloader.utils.showToast
import com.google.android.material.dialog.MaterialAlertDialogBuilder
import kotlinx.coroutines.launch

class HomeFragment : Fragment() {

    private var _binding: FragmentHomeBinding? = null
    private val binding get() = _binding!!

    private val viewModel: HomeViewModel by viewModels()
    private lateinit var platformAdapter: PlatformAdapter

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

        setupRecyclerView()
        setupListeners()
        observeViewModel()

        viewModel.loadPlatforms()
    }

    private fun setupRecyclerView() {
        platformAdapter = PlatformAdapter { platform ->
            // Handle platform click - auto-fill platform info
            requireContext().showToast("Selected: ${platform.name}")
        }

        binding.rvPlatforms.apply {
            adapter = platformAdapter
            layoutManager = GridLayoutManager(requireContext(), 2)
        }
    }

    private fun setupListeners() {
        binding.btnDownload.setOnClickListener {
            val url = binding.etUrl.text.toString().trim()

            if (url.isEmpty()) {
                binding.tilUrl.error = "Please enter a URL"
                return@setOnClickListener
            }

            if (!url.startsWith("http")) {
                binding.tilUrl.error = "Please enter a valid URL"
                return@setOnClickListener
            }

            binding.tilUrl.error = null
            viewModel.createDownload(url)
        }

        binding.etUrl.setOnFocusChangeListener { _, hasFocus ->
            if (!hasFocus) {
                binding.tilUrl.error = null
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
                        requireContext().showToast(resource.message ?: "Failed to load platforms")
                    }
                }
            }
        }

        viewLifecycleOwner.lifecycleScope.launch {
            viewModel.downloadResult.collect { resource ->
                when (resource) {
                    is Resource.Loading -> {
                        binding.btnDownload.isEnabled = false
                        binding.progressBar.visibility = View.VISIBLE
                    }
                    is Resource.Success -> {
                        binding.btnDownload.isEnabled = true
                        binding.progressBar.visibility = View.GONE

                        resource.data?.let { task ->
                            // Show format selection dialog if formats available
                            if (!task.formats.isNullOrEmpty()) {
                                showFormatSelectionDialog(task)
                            } else {
                                requireContext().showToast("Download started!")
                                binding.etUrl.text?.clear()
                            }
                        }
                    }
                    is Resource.Error -> {
                        binding.btnDownload.isEnabled = true
                        binding.progressBar.visibility = View.GONE
                        requireContext().showToast(resource.message ?: "Download failed")
                    }
                }
            }
        }
    }

    private fun showFormatSelectionDialog(task: com.agcforge.videodownloader.data.model.DownloadTask) {
        val formats = task.formats ?: return
        val formatNames = formats.map { it.getFormatDescription() }.toTypedArray()

        MaterialAlertDialogBuilder(requireContext())
            .setTitle("Select Quality")
            .setItems(formatNames) { dialog, which ->
                val selectedFormat = formats[which]
                requireContext().showToast("Downloading ${selectedFormat.getQualityLabel()}")
                binding.etUrl.text?.clear()
                dialog.dismiss()
            }
            .setNegativeButton("Cancel", null)
            .show()
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}