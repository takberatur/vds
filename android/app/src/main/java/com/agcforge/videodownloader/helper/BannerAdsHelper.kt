package com.agcforge.videodownloader.helper

import android.app.Activity
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.FrameLayout
import com.agcforge.videodownloader.R
import com.google.android.gms.ads.AdListener
import com.google.android.gms.ads.AdRequest
import com.google.android.gms.ads.AdView
import com.google.android.gms.ads.LoadAdError
import com.startapp.sdk.ads.banner.Banner
import com.startapp.sdk.ads.banner.BannerListener
import com.unity3d.services.banners.BannerErrorInfo
import com.unity3d.services.banners.BannerView
import com.unity3d.services.banners.UnityBannerSize

class BannerAdsHelper(private val activity: Activity) {

    private val TAG = "BannerAdsHelper"

    private var admobBannerView: AdView? = null
    private var unityBannerView: BannerView? = null
    private var startIoBanner: Banner? = null

    private var admobContainer: View? = null
    private var unityContainer: FrameLayout? = null
    private var startIoContainer: FrameLayout? = null

    // Loading states
    private var isAdmobLoading = false
    private var isUnityLoading = false
    private var isStartIoLoading = false

    // Current provider
    private var currentProvider: AdsConfig.AdsProvider? = null

    fun inflateBannerView(parent: ViewGroup): View {
        return LayoutInflater.from(activity)
            .inflate(R.layout.item_banner_ad, parent, false)
    }

    fun loadAndAttachBanner(
        container: ViewGroup,
        onBannerLoaded: ((AdsConfig.AdsProvider) -> Unit)? = null,
        onBannerFailed: ((AdsConfig.AdsProvider) -> Unit)? = null
    ) {
        if (!AdsConfig.ENABLE_ADS) {
            Log.d(TAG, "Ads disabled")
            container.visibility = View.GONE
            return
        }

        // Inflate banner layout
        val bannerView = inflateBannerView(container)
        container.addView(bannerView)

        // Get references
        admobContainer = bannerView.findViewById(R.id.admobBannerView)
        unityContainer = bannerView.findViewById(R.id.unityBannerContainer)
        startIoContainer = bannerView.findViewById(R.id.startIoBannerContainer)

        admobBannerView = admobContainer as? AdView

        // Load banner
        loadBanner(onBannerLoaded, onBannerFailed)
    }

    fun loadBanner(
        onBannerLoaded: ((AdsConfig.AdsProvider) -> Unit)? = null,
        onBannerFailed: ((AdsConfig.AdsProvider) -> Unit)? = null
    ) {
        if (!AdsConfig.ENABLE_ADS) {
            Log.d(TAG, "Ads disabled")
            return
        }

        // Try providers in priority order
        for (provider in AdsConfig.Rotation.BANNER_PRIORITY) {
            when (provider) {
                AdsConfig.AdsProvider.ADMOB -> {
                    if (admobBannerView != null && !isAdmobLoading) {
                        loadAdmobBanner(onBannerLoaded, onBannerFailed)
                        return
                    }
                }
                AdsConfig.AdsProvider.UNITY -> {
                    if (unityContainer != null && !isUnityLoading) {
                        loadUnityBanner(onBannerLoaded, onBannerFailed)
                        return
                    }
                }
                AdsConfig.AdsProvider.STARTIO -> {
                    if (startIoContainer != null && !isStartIoLoading) {
                        loadStartIoBanner(onBannerLoaded, onBannerFailed)
                        return
                    }
                }
            }
        }
    }


    private fun loadAdmobBanner(
        onBannerLoaded: ((AdsConfig.AdsProvider) -> Unit)?,
        onBannerFailed: ((AdsConfig.AdsProvider) -> Unit)?
    ) {
        admobBannerView?.apply {
            isAdmobLoading = true

            adListener = object : AdListener() {
                override fun onAdLoaded() {
                    Log.d(TAG, "Admob banner loaded")
                    isAdmobLoading = false
                    currentProvider = AdsConfig.AdsProvider.ADMOB
                    showBanner(AdsConfig.AdsProvider.ADMOB)
                    onBannerLoaded?.invoke(AdsConfig.AdsProvider.ADMOB)
                }

                override fun onAdFailedToLoad(error: LoadAdError) {
                    Log.e(TAG, "Admob banner failed: ${error.message}")
                    isAdmobLoading = false
                    tryNextProvider(AdsConfig.AdsProvider.ADMOB, onBannerLoaded, onBannerFailed)
                }

                override fun onAdClicked() {
                    Log.d(TAG, "Admob banner clicked")
                }
            }

            loadAd(AdRequest.Builder().build())
        }
    }

    private fun loadUnityBanner(
        onBannerLoaded: ((AdsConfig.AdsProvider) -> Unit)?,
        onBannerFailed: ((AdsConfig.AdsProvider) -> Unit)?
    ) {
        unityContainer?.apply {
            isUnityLoading = true
            removeAllViews()

            unityBannerView = BannerView(
                activity,
                AdsConfig.unityConfig.bannerPlacement,
                UnityBannerSize(320, 50)
            ).apply {
                listener = object : BannerView.IListener {
                    override fun onBannerLoaded(bannerView: BannerView?) {
                        Log.d(TAG, "Unity banner loaded")
                        isUnityLoading = false
                        currentProvider = AdsConfig.AdsProvider.UNITY
                        showBanner(AdsConfig.AdsProvider.UNITY)
                        onBannerLoaded?.invoke(AdsConfig.AdsProvider.UNITY)
                    }

                    override fun onBannerShown(bannerAdView: BannerView?) {}

                    override fun onBannerFailedToLoad(
                        bannerView: BannerView?,
                        errorInfo: BannerErrorInfo?
                    ) {
                        Log.e(TAG, "Unity banner failed: ${errorInfo?.errorMessage}")
                        isUnityLoading = false
                        tryNextProvider(AdsConfig.AdsProvider.UNITY, onBannerLoaded, onBannerFailed)
                    }

                    override fun onBannerClick(bannerView: BannerView?) {
                        Log.d(TAG, "Unity banner clicked")
                    }

                    override fun onBannerLeftApplication(bannerView: BannerView?) {
                        Log.d(TAG, "Unity banner left application")
                    }
                }
                load()
            }

            addView(unityBannerView)
        }
    }

    private fun loadStartIoBanner(
        onBannerLoaded: ((AdsConfig.AdsProvider) -> Unit)?,
        onBannerFailed: ((AdsConfig.AdsProvider) -> Unit)?
    ) {
        startIoContainer?.apply {
            isStartIoLoading = true
            removeAllViews()

            startIoBanner = Banner(activity, object : BannerListener {
                override fun onReceiveAd(view: View) {
                    Log.d(TAG, "Start.io banner loaded")
                    isStartIoLoading = false
                    currentProvider = AdsConfig.AdsProvider.STARTIO
                    showBanner(AdsConfig.AdsProvider.STARTIO)
                    onBannerLoaded?.invoke(AdsConfig.AdsProvider.STARTIO)
                }

                override fun onFailedToReceiveAd(view: View?) {
                    Log.e(TAG, "Start.io banner failed to load.")
                    isStartIoLoading = false
                    tryNextProvider(AdsConfig.AdsProvider.STARTIO, onBannerLoaded, onBannerFailed)
                }

                override fun onImpression(view: View) {}
                override fun onClick(view: View) {}
            })

            addView(startIoBanner)
        }
    }

    private fun showBanner(provider: AdsConfig.AdsProvider) {
        admobContainer?.visibility = View.GONE
        unityContainer?.visibility = View.GONE
        startIoContainer?.visibility = View.GONE

        when (provider) {
            AdsConfig.AdsProvider.ADMOB -> admobContainer?.visibility = View.VISIBLE
            AdsConfig.AdsProvider.UNITY -> unityContainer?.visibility = View.VISIBLE
            AdsConfig.AdsProvider.STARTIO -> startIoContainer?.visibility = View.VISIBLE
        }
    }

    private fun tryNextProvider(
        failedProvider: AdsConfig.AdsProvider,
        onBannerLoaded: ((AdsConfig.AdsProvider) -> Unit)?,
        onBannerFailed: ((AdsConfig.AdsProvider) -> Unit)?
    ) {
        val currentIndex = AdsConfig.Rotation.BANNER_PRIORITY.indexOf(failedProvider)
        val nextProviders = AdsConfig.Rotation.BANNER_PRIORITY.drop(currentIndex + 1)

        for (provider in nextProviders) {
            when (provider) {
                AdsConfig.AdsProvider.ADMOB -> {
                    if (admobBannerView != null && !isAdmobLoading) {
                        loadAdmobBanner(onBannerLoaded, onBannerFailed)
                        return
                    }
                }
                AdsConfig.AdsProvider.UNITY -> {
                    if (unityContainer != null && !isUnityLoading) {
                        loadUnityBanner(onBannerLoaded, onBannerFailed)
                        return
                    }
                }
                AdsConfig.AdsProvider.STARTIO -> {
                    if (startIoContainer != null && !isStartIoLoading) {
                        loadStartIoBanner(onBannerLoaded, onBannerFailed)
                        return
                    }
                }
            }
        }

        Log.e(TAG, "All banner providers failed")
        hideBanner()
        onBannerFailed?.invoke(failedProvider)
    }

    fun hideBanner() {
        admobContainer?.visibility = View.GONE
        unityContainer?.visibility = View.GONE
        startIoContainer?.visibility = View.GONE
    }

    fun showCurrentBanner() {
        currentProvider?.let { showBanner(it) }
    }

    fun destroy() {
        admobBannerView?.destroy()
        unityBannerView?.destroy()
        startIoBanner = null

        admobBannerView = null
        unityBannerView = null
    }

    fun pause() {
        admobBannerView?.pause()
    }

    fun resume() {
        admobBannerView?.resume()
    }
}
