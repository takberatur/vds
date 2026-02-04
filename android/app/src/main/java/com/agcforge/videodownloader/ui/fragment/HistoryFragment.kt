package com.agcforge.videodownloader.ui.fragment

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.Fragment
import androidx.lifecycle.lifecycleScope
import androidx.navigation.fragment.findNavController
import androidx.recyclerview.widget.LinearLayoutManager
import com.agcforge.videodownloader.databinding.FragmentHistoryBinding
import com.agcforge.videodownloader.ui.adapter.HistoryAdapter
import com.agcforge.videodownloader.utils.PreferenceManager
import kotlinx.coroutines.launch

class HistoryFragment : Fragment() {

    private var _binding: FragmentHistoryBinding? = null
    private val binding get() = _binding!!

    private lateinit var preferenceManager: PreferenceManager
    private lateinit var historyAdapter: HistoryAdapter

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = FragmentHistoryBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        preferenceManager = PreferenceManager(requireContext())

        setupRecyclerView()
        observeHistory()

        binding.btnClearHistory.setOnClickListener {
            lifecycleScope.launch {
                preferenceManager.clearHistory()
            }
        }
    }

    private fun setupRecyclerView() {
        historyAdapter = HistoryAdapter { url ->
            // Pass the selected URL back to HomeFragment
            findNavController().previousBackStackEntry?.savedStateHandle?.set("history_url", url)
            findNavController().popBackStack()
        }

        binding.rvHistory.apply {
            adapter = historyAdapter
            layoutManager = LinearLayoutManager(requireContext())
        }
    }

    private fun observeHistory() {
        lifecycleScope.launch {
            preferenceManager.history.collect { historyList ->
                val isHistoryEmpty = historyList.isEmpty()
                binding.tvEmpty.visibility = if (isHistoryEmpty) View.VISIBLE else View.GONE
                binding.rvHistory.visibility = if (isHistoryEmpty) View.GONE else View.VISIBLE
                binding.btnClearHistory.visibility = if (isHistoryEmpty) View.GONE else View.VISIBLE

                historyAdapter.submitList(historyList.reversed()) // Show the most recent items at the top
            }
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
